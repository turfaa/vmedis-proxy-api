package sale

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Service struct {
	db           *Database
	vmedis       *vmedis.Client
	drugsGetter  DrugsGetter
	drugProducer UpdatedDrugProducer
}

func (s *Service) GetSalesBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]Sale, error) {
	return s.db.GetSalesBetweenTime(ctx, from, to)
}

func (s *Service) GetAggregatedSalesBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]AggregatedSale, error) {
	return s.db.GetAggregatedSalesBetweenTime(ctx, from, to)
}

func (s *Service) GetSoldDrugsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]SoldDrug, error) {
	sales, err := s.db.GetSalesBetweenTime(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get sales between %s - %s: %w", from, to, err)
	}

	var (
		drugCodes []string
		soldDrugs = make(map[string]SoldDrug)
	)
	for _, sale := range sales {
		for _, saleUnit := range sale.SaleUnits {
			drugCodes = append(drugCodes, saleUnit.DrugCode)
			soldDrugs[saleUnit.DrugCode] = SoldDrug{
				Occurrences: soldDrugs[saleUnit.DrugCode].Occurrences + 1,
				TotalAmount: soldDrugs[saleUnit.DrugCode].TotalAmount + saleUnit.Total,
			}
		}
	}

	drugs, err := s.drugsGetter.GetDrugsByVmedisCodes(ctx, drugCodes)
	if err != nil {
		return nil, fmt.Errorf("get drugs by vmedis codes: %w", err)
	}

	for _, d := range drugs {
		soldDrugs[d.VmedisCode] = SoldDrug{
			Drug:        d,
			Occurrences: soldDrugs[d.VmedisCode].Occurrences,
			TotalAmount: soldDrugs[d.VmedisCode].TotalAmount,
		}
	}

	soldDrugsSlice := make([]SoldDrug, 0, len(soldDrugs))
	for _, d := range soldDrugs {
		soldDrugsSlice = append(soldDrugsSlice, d)
	}

	sort.Slice(soldDrugsSlice, func(i, j int) bool {
		if soldDrugsSlice[i].Occurrences == soldDrugsSlice[j].Occurrences {
			return soldDrugsSlice[i].TotalAmount > soldDrugsSlice[j].TotalAmount
		}

		return soldDrugsSlice[i].Occurrences > soldDrugsSlice[j].Occurrences
	})

	return soldDrugsSlice, nil
}

func (s *Service) GetSalesStatisticsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]Statistics, error) {
	stats, err := s.db.GetSalesStatisticsBetweenTime(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get sales statistics between %s - %s from DB: %w", from, to, err)
	}

	if to.Before(time.Now()) {
		return stats, nil
	}

	liveStatsVmedis, err := s.vmedis.GetDailySalesStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("get live sales statistics from vmedis: %w", err)
	}

	liveStats, err := FromVmedisSalesStatistics(time.Now(), liveStatsVmedis)
	if err != nil {
		return nil, fmt.Errorf("invalid live sales statistics from Vmedis (%+v): %w", liveStatsVmedis, err)
	}

	if len(stats) == 0 || (stats[len(stats)-1].TotalSales <= liveStats.TotalSales && stats[len(stats)-1].PulledAt.Before(liveStats.PulledAt)) {
		stats = append(stats, liveStats)
	}

	return stats, nil
}

func (s *Service) DumpTodaySalesFromVmedisToDB(ctx context.Context) error {
	log.Printf("Dumping today's sales from Vmedis to DB")

	vmedisSales, err := s.vmedis.GetAllTodaySales(ctx)
	if err != nil {
		return fmt.Errorf("get sales from vmedis: %w", err)
	}

	log.Printf("Got %d sales from Vmedis, dumping to DB", len(vmedisSales))

	if err := s.db.UpsertVmedisSales(ctx, vmedisSales); err != nil {
		return fmt.Errorf("upsert sales to DB: %w", err)
	}

	log.Printf("Dumped today's sales from Vmedis to DB, producing updated drug messages")

	messages := make([]*kafkapb.UpdatedDrugByVmedisCode, 0, len(vmedisSales))
	for _, sale := range vmedisSales {
		for _, saleUnit := range sale.SaleUnits {
			messages = append(messages, &kafkapb.UpdatedDrugByVmedisCode{
				RequestKey: fmt.Sprintf("sale:%s:%s", sale.InvoiceNumber, saleUnit.DrugCode),
				VmedisCode: saleUnit.DrugCode,
			})
		}
	}

	if err := s.drugProducer.ProduceUpdatedDrugByVmedisCode(ctx, messages); err != nil {
		return fmt.Errorf("produce updated drug messages: %w", err)
	}

	log.Printf("Produced updated drug messages, finished dumping today's sales from Vmedis to DB")

	return nil
}

func (s *Service) DumpTodaySalesStatisticsFromVmedisToDB(ctx context.Context) error {
	log.Println("Dumping today's sales statistics from Vmedis to DB")

	vmedisStats, err := s.vmedis.GetDailySalesStatistics(ctx)
	if err != nil {
		return fmt.Errorf("get sales statistics from vmedis: %w", err)
	}

	stats, err := FromVmedisSalesStatistics(time.Now(), vmedisStats)
	if err != nil {
		return fmt.Errorf("invalid sales statistics from Vmedis (%+v): %w", vmedisStats, err)
	}

	if err := s.db.InsertSalesStatistics(ctx, stats); err != nil {
		return fmt.Errorf("insert sales statistics to DB: %w", err)
	}

	log.Printf("Dumped today's sales statistics from Vmedis to DB (total sales: %.2f, number of sales: %d)", stats.TotalSales, stats.NumberOfSales)
	return nil
}

func NewService(
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) *Service {
	return &Service{
		db:           NewDatabase(db),
		vmedis:       vmedisClient,
		drugsGetter:  drugsGetter,
		drugProducer: drugProducer,
	}
}
