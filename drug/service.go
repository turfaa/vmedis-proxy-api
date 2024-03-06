package drug

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	"github.com/turfaa/vmedis-proxy-api/pkg2/chans"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
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
	db       *Database
	vmedis   *vmedis.Client
	producer *Producer
}

// GetDrugs returns all drugs.
func (s *Service) GetDrugs(ctx context.Context) ([]Drug, error) {
	return getDrugsFromDB(ctx, s.db.GetDrugsUpdatedAfter, 1000)
}

func (s *Service) GetDrugsByVmedisCodes(ctx context.Context, vmedisCodes []string) ([]Drug, error) {
	return getDrugsFromDB(ctx, func(ctx context.Context, minimumUpdatedTime time.Time) ([]models.Drug, error) {
		return s.db.GetDrugsByVmedisCodesUpdatedAfter(ctx, vmedisCodes, minimumUpdatedTime)
	}, len(vmedisCodes))
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

	requestKey := fmt.Sprintf("dump_drugs_from_vmedis_to_db:%s", time.Now().Format("2006-01-02_15-04-05"))

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

		updatedDrugs := make([]*kafkapb.UpdatedDrugByVmedisID, 0, len(batch))
		for _, drug := range batch {
			updatedDrugs = append(updatedDrugs, &kafkapb.UpdatedDrugByVmedisID{
				RequestKey: fmt.Sprintf("%s:%d", requestKey, drug.VmedisID),
				VmedisId:   drug.VmedisID,
			})
		}

		log.Printf("Producing %d %s messages [batch %d]", len(updatedDrugs), VmedisIDUpdatedTopic, batchNum)

		if err := s.producer.ProduceUpdatedDrugsByVmedisID(ctx, updatedDrugs); err != nil {
			log.Printf("Error producing %d %s messages [batch %d]: %s", len(updatedDrugs), VmedisIDUpdatedTopic, batchNum, err)
			errs = append(errs, err)
		} else {
			log.Printf("Produced %d %s messages [batch %d]", len(updatedDrugs), VmedisIDUpdatedTopic, batchNum)
		}

		log.Printf("Finished dumping drugs batch %d, number of drugs: %d", batchNum, len(batch))
		batchNum++
	}

	log.Printf("Finished dumping drugs, got %d errors, errs: %s", len(errs), errors.Join(errs...))
	return nil
}

func (s *Service) DumpDrugDetailsFromVmedisToDBByVmedisCode(ctx context.Context, vmedisCode string) error {
	log.Printf("Starting to dump drug details of %s from Vmedis to DB", vmedisCode)

	log.Printf("Getting drug %s from DB", vmedisCode)
	drugs, err := s.db.GetDrugsByVmedisCodesUpdatedAfter(ctx, []string{vmedisCode}, time.Now().Add(-drugsUpdatedAtThresholds[0]))
	if err != nil {
		return fmt.Errorf("get drug %s from DB: %w", vmedisCode, err)
	}
	if len(drugs) == 0 {
		return fmt.Errorf("drug %s not found in DB", vmedisCode)
	}

	drug := drugs[0]

	return s.DumpDrugDetailsFromVmedisToDBByVmedisID(ctx, drug.VmedisID)
}

// DumpDrugDetailsFromVmedisToDBByVmedisID dumps the details of a drug from Vmedis to DB.
func (s *Service) DumpDrugDetailsFromVmedisToDBByVmedisID(ctx context.Context, vmedisID int64) error {
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
func NewService(db *gorm.DB, vmedisClient *vmedis.Client, kafkaWriter *kafka.Writer) *Service {
	return &Service{
		db:       NewDatabase(db),
		vmedis:   vmedisClient,
		producer: NewProducer(kafkaWriter),
	}
}
