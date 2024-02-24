package stockopname

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Service struct {
	db           *Database
	vmedis       *vmedis.Client
	drugProducer UpdatedDrugProducer
}

func (s *Service) GetStockOpnamesBetweenTime(ctx context.Context, from, to time.Time) ([]StockOpname, error) {
	return s.db.GetStockOpnamesBetweenTime(ctx, from, to)
}

func (s *Service) GetCompactedStockOpnamesBetweenTime(ctx context.Context, from, to time.Time) ([]CompactedStockOpname, error) {
	stockOpnames, err := s.db.GetStockOpnamesBetweenTime(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get stock opnames from DB: %w", err)
	}

	compacted := make([]CompactedStockOpname, 0, len(stockOpnames))

	var currentDrug []StockOpname
	for _, so := range stockOpnames {
		if len(currentDrug) == 0 || stockOpnameShouldCompact(so, currentDrug[0]) {
			currentDrug = append(currentDrug, so)
			continue
		}

		compacted = append(compacted, compactOneDrugStockOpnames(currentDrug))
		currentDrug = append(currentDrug[:0], so)
	}

	if len(currentDrug) > 0 {
		compacted = append(compacted, compactOneDrugStockOpnames(currentDrug))
	}

	return compacted, nil
}

func stockOpnameShouldCompact(so1, so2 StockOpname) bool {
	return so1.Date == so2.Date && so1.DrugCode == so2.DrugCode && so1.Unit == so2.Unit
}

// compactOneDrugStockOpnames assumes that the stock opnames are:
// - In the same date
// - Sorted chronologically
// - For the same drug
func compactOneDrugStockOpnames(stockOpnames []StockOpname) CompactedStockOpname {
	compacted := CompactedStockOpname{
		Date:     stockOpnames[0].Date,
		DrugCode: stockOpnames[0].DrugCode,
		DrugName: stockOpnames[0].DrugName,
		Unit:     stockOpnames[0].Unit,
		Changes:  make([]StockChange, 0, len(stockOpnames)),
	}

	for _, so := range stockOpnames {
		if so.InitialQuantity == so.RealQuantity {
			continue
		}

		compacted.QuantityDifference += so.QuantityDifference
		compacted.HPPDifference += so.HPPDifference
		compacted.SalePriceDifference += so.SalePriceDifference

		compacted.Changes = append(compacted.Changes, StockChange{
			BatchCode:       so.BatchCode,
			InitialQuantity: so.InitialQuantity,
			RealQuantity:    so.RealQuantity,
		})
	}

	return compacted
}

func (s *Service) GetStockOpnameSummariesBetweenTime(ctx context.Context, from, to time.Time) ([]Summary, error) {
	stockOpnames, err := s.db.GetStockOpnamesBetweenTime(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get stock opnames from DB: %w", err)
	}

	summaryMap := make(map[string]Summary, len(stockOpnames))
	for _, so := range stockOpnames {
		key := fmt.Sprintf("%s#%s", so.DrugCode, so.Unit)
		summaryMap[key] = addStockOpnameToSummary(summaryMap[key], so)
	}

	summaries := make([]Summary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func addStockOpnameToSummary(summary Summary, stockOpname StockOpname) Summary {
	summary.DrugCode = stockOpname.DrugCode
	summary.DrugName = stockOpname.DrugName
	summary.Unit = stockOpname.Unit
	summary.Changes = append(summary.Changes, StockChange{
		BatchCode:       stockOpname.BatchCode,
		InitialQuantity: stockOpname.InitialQuantity,
		RealQuantity:    stockOpname.RealQuantity,
	})
	summary.QuantityDifference += stockOpname.QuantityDifference
	summary.HPPDifference += stockOpname.HPPDifference
	summary.SalePriceDifference += stockOpname.SalePriceDifference
	return summary
}

func (s *Service) DumpTodayStockOpnamesFromVmedisToDB(ctx context.Context) error {
	log.Println("Dumping stock opnames")

	log.Println("Getting all today's stock opnames from Vmedis")
	stockOpnames, err := s.vmedis.GetAllTodayStockOpnames(ctx)
	if err != nil {
		return fmt.Errorf("get all today's stock opnames from Vmedis: %w", err)
	}

	if len(stockOpnames) == 0 {
		log.Println("No stock opnames found")
		return nil
	}

	log.Printf("Got %d stock opnames from Vmedis", len(stockOpnames))

	log.Println("Upserting stock opnames to DB")
	if err := s.db.UpsertVmedisStockOpnames(ctx, stockOpnames); err != nil {
		return fmt.Errorf("upsert vmedis stock opnames: %w", err)
	}
	log.Println("Done upserting stock opnames to DB")

	kafkaMessages := make([]*kafkapb.UpdatedDrugByVmedisCode, 0, len(stockOpnames))
	for _, so := range stockOpnames {
		kafkaMessages = append(kafkaMessages, &kafkapb.UpdatedDrugByVmedisCode{
			RequestKey: fmt.Sprintf("stock-opname:%s:%s", so.ID, so.DrugCode),
			VmedisCode: so.DrugCode,
		})
	}

	log.Printf("Producing %d updated drugs kafka messages", len(kafkaMessages))
	if err := s.drugProducer.ProduceUpdatedDrugByVmedisCode(ctx, kafkaMessages); err != nil {
		log.Printf("Error producing updated drugs: %s", err)
	} else {
		log.Printf("Produced %d updated drugs kafka messages", len(kafkaMessages))
	}

	log.Println("Done dumping stock opnames")
	return nil
}

func NewService(db *gorm.DB, vmedisClient *vmedis.Client, drugProducer UpdatedDrugProducer) *Service {
	return &Service{
		db:           NewDatabase(db),
		vmedis:       vmedisClient,
		drugProducer: drugProducer,
	}
}
