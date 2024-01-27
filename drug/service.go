package drug

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/chans"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

const (
	insertBatchSize = 20
)

var (
	drugsUpdatedAtThresholds = []time.Duration{
		24*time.Hour + 3*time.Hour,
		3 * 24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
		9999 * 24 * time.Hour,
	}
)

// Service provides business logic related to drugs.
type Service struct {
	db     *Database
	vmedis *vmedis.Client
}

// GetDrugs returns all drugs.
func (s *Service) GetDrugs(ctx context.Context) ([]Drug, error) {
	return getDrugsFromDB(ctx, s.db.GetDrugsUpdatedAfter, 1000)
}

// GetDrugsToStockOpname returns the drugs to stock opname.
// Drugs that have been already stock opnamed between the given times will be excluded.
func (s *Service) GetDrugsToStockOpname(ctx context.Context, startTime time.Time, endTime time.Time) ([]Drug, error) {
	alreadyStockOpnamedDrugCodes, err := s.db.GetDrugCodesAlreadyStockOpnamedBetweenTimes(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get drug codes already stock opnamed between %s and %s from DB: %w", startTime, endTime, err)
	}

	drugs, err := getDrugsFromDB(
		ctx,
		func(ctx context.Context, minimumUpdatedTime time.Time) ([]models.Drug, error) {
			return s.db.GetDrugsByExcludedVmedisCodesUpdatedAfter(ctx, alreadyStockOpnamedDrugCodes, minimumUpdatedTime)
		},
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("get drugs except %v from DB: %w", alreadyStockOpnamedDrugCodes, err)
	}

	return drugs, nil
}

// GetSalesBasedDrugsToStockOpname returns the drugs to stock opname based on sales.
func (s *Service) GetSalesBasedDrugsToStockOpname(ctx context.Context, startTime time.Time, endTime time.Time) ([]Drug, error) {
	drugSaleStatistics, err := s.db.GetDrugSaleStatisticsBetweenTimes(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get drug sale statistics between %s and %s from DB: %w", startTime, endTime, err)
	}

	alreadyStockOpnamedDrugCodes, err := s.db.GetDrugCodesAlreadyStockOpnamedBetweenTimes(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get drug codes already stock opnamed between %s and %s from DB: %w", startTime, endTime, err)
	}

	drugsToStockOpnameCodesMap := make(map[string]struct{}, len(drugSaleStatistics))
	saleStatisticsByDrugCode := make(map[string]SaleStatistics, len(drugSaleStatistics))
	for _, stats := range drugSaleStatistics {
		drugsToStockOpnameCodesMap[stats.DrugCode] = struct{}{}
		saleStatisticsByDrugCode[stats.DrugCode] = stats
	}

	for _, drugCode := range alreadyStockOpnamedDrugCodes {
		delete(drugsToStockOpnameCodesMap, drugCode)
	}

	drugsToStockOpnameCodes := make([]string, 0, len(drugsToStockOpnameCodesMap))
	for drugCode := range drugsToStockOpnameCodesMap {
		drugsToStockOpnameCodes = append(drugsToStockOpnameCodes, drugCode)
	}

	drugs, err := getDrugsFromDB(
		ctx,
		func(ctx context.Context, minimumUpdatedTime time.Time) ([]models.Drug, error) {
			return s.db.GetDrugsByVmedisCodesUpdatedAfter(ctx, drugsToStockOpnameCodes, minimumUpdatedTime)
		},
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("get drugs by vmedis codes updated after %s from DB: %w", startTime, err)
	}

	sortDrugsByNumberOfSales(drugs, saleStatisticsByDrugCode)
	return drugs, nil
}

type dbDrugsGetter func(ctx context.Context, minimumUpdatedTime time.Time) ([]models.Drug, error)

func getDrugsFromDB(ctx context.Context, getter dbDrugsGetter, minimumDrugsSize int) ([]Drug, error) {
	var drugs []Drug

	for _, threshold := range drugsUpdatedAtThresholds {
		dbDrugs, err := getter(ctx, time.Now().Add(-threshold))
		if err != nil {
			return nil, fmt.Errorf("get drugs updated after %s: %w", threshold, err)
		}

		drugs = slices2.Map(dbDrugs, FromDBDrug)

		if len(drugs) >= minimumDrugsSize {
			break
		}
	}

	return drugs, nil
}

func sortDrugsByNumberOfSales(drugs []Drug, saleStatisticsByDrugCode map[string]SaleStatistics) {
	slices.SortFunc(drugs, func(i, j Drug) int {
		iSaleStats := saleStatisticsByDrugCode[i.VmedisCode]
		jSaleStats := saleStatisticsByDrugCode[j.VmedisCode]

		if iSaleStats.NumberOfSales == jSaleStats.NumberOfSales {
			return int(jSaleStats.TotalAmount*100 - iSaleStats.TotalAmount*100)
		}

		return jSaleStats.NumberOfSales - iSaleStats.NumberOfSales
	})
}

// DumpDrugsFromVmedisToDB dumps the drugs from Vmedis to DB.
func (s *Service) DumpDrugsFromVmedisToDB(ctx context.Context) error {
	log.Println("Dumping drugs from Vmedis to DB")

	drugs, err := s.vmedis.GetAllDrugs(ctx)
	if err != nil {
		return fmt.Errorf("get all drugs from Vmedis: %w", err)
	}

	log.Printf("Got %d drugs from Vmedis", len(drugs))

	batches := chans.GenerateBatches(drugs, insertBatchSize)

	var errs []error
	batchNum := 1
	for batch := range batches {
		log.Printf("Starting to dump drugs batch %d, number of drugs: %d", batchNum, len(batch))

		log.Println("Upserting drugs to DB")
		if err := s.db.UpsertVmedisDrugs(ctx, batch, "vmedis_code", []string{"vmedis_id", "name", "manufacturer"}); err != nil {
			log.Printf("Error upserting drugs to DB: %s", err)
			errs = append(errs, err)
			continue
		}
		log.Println("Upserted drugs to DB")

		for _, drug := range batch {
			log.Printf("Dumping drug %d details to DB", drug.VmedisID)
			if err := s.DumpDrugDetailsFromVmedisToDB(ctx, drug.VmedisID); err != nil {
				log.Printf("Error dumping drug %d details to DB: %s", drug.VmedisID, err)
				errs = append(errs, err)
				continue
			}
			log.Printf("Dumped drug %d details to DB", drug.VmedisID)
		}

		log.Printf("Finished dumping drugs batch %d, number of drugs: %d", batchNum, len(batch))
		batchNum++
	}

	log.Printf("Finished dumping drugs, got %d errors, errs: %s", len(errs), errors.Join(errs...))
	return nil
}

// DumpDrugDetailsFromVmedisToDB dumps the details of a drug from Vmedis to DB.
func (s *Service) DumpDrugDetailsFromVmedisToDB(ctx context.Context, vmedisID int) error {
	log.Printf("Starting to dump drug details of %d from Vmedis to DB", vmedisID)

	log.Printf("Getting drug %d from Vmedis", vmedisID)
	drug, err := s.vmedis.GetDrug(ctx, vmedisID)
	if err != nil {
		return fmt.Errorf("get drug %d from Vmedis: %w", vmedisID, err)
	}
	log.Printf("Got drug %d from Vmedis", vmedisID)

	log.Printf("Upserting drug %d details to DB", vmedisID)
	if err := s.db.UpsertVmedisDrug(
		ctx,
		drug,
		"vmedis_id",
		[]string{"vmedis_code", "name", "manufacturer", "minimum_stock_unit", "minimum_stock_quantity"},
	); err != nil {
		return fmt.Errorf("upsert drug %d details to DB: %w", vmedisID, err)
	}
	log.Printf("Upserted drug %d details to DB", vmedisID)

	log.Printf("Upserting %d drug %d units to DB", len(drug.Units), vmedisID)
	if err := s.db.UpsertVmedisDrugUnits(ctx, drug.VmedisCode, drug.Units); err != nil {
		return fmt.Errorf("upsert drug %d units to DB: %w", vmedisID, err)
	}
	log.Printf("Upserted %d drug %d units to DB", len(drug.Units), vmedisID)

	log.Printf("Upserting %d drug %d stocks to DB", len(drug.Stocks), vmedisID)
	if err := s.db.UpsertVmedisDrugStocks(ctx, drug.VmedisCode, drug.Stocks); err != nil {
		return fmt.Errorf("upsert drug %d stocks to DB: %w", vmedisID, err)
	}
	log.Printf("Upserted %d drug %d stocks to DB", len(drug.Stocks), vmedisID)

	log.Printf("Finished dumping drug details of %d from Vmedis to DB", vmedisID)
	return nil
}

// NewService creates a new drug service.
func NewService(db *gorm.DB, vmedisClient *vmedis.Client) *Service {
	return &Service{
		db:     NewDatabase(db),
		vmedis: vmedisClient,
	}
}
