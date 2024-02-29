package procurement

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Service struct {
	db              *Database
	redisDB         *RedisDatabase
	vmedis          *vmedis.Client
	drugProducer    UpdatedDrugProducer
	drugUnitsGetter DrugUnitsGetter
}

func (s *Service) GetAggregatedProcurementsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]AggregatedProcurement, error) {
	return s.db.GetAggregatedProcurementsBetweenTime(ctx, from, to)
}

func (s *Service) GetRecommendations(ctx context.Context) (RecommendationsResponse, error) {
	recommendations, err := s.redisDB.GetRecommendations(ctx)
	if err != nil {
		return RecommendationsResponse{}, fmt.Errorf("get procurement recommendations from Redis: %w", err)
	}

	return recommendations, nil
}

func (s *Service) DumpProcurementsBetweenDatesFromVmedisToDB(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
) error {
	log.Printf("Dumping procurements from %s to %s from vmedis to DB", startDate, endDate)

	procurements, err := s.vmedis.GetAllProcurementsBetweenDates(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("get all procurements between %s and %s from vmedis: %w", startDate, endDate, err)
	}

	log.Printf("Got %d procurements from vmedis", len(procurements))

	if err := s.db.UpsertVmedisProcurements(ctx, procurements); err != nil {
		return fmt.Errorf("upsert vmedis procurements: %w", err)
	}

	log.Printf("Dumped %d procurements from vmedis to DB", len(procurements))

	var updatedDrugs []*kafkapb.UpdatedDrugByVmedisCode
	for _, p := range procurements {
		for _, pu := range p.ProcurementUnits {
			updatedDrugs = append(updatedDrugs, &kafkapb.UpdatedDrugByVmedisCode{
				RequestKey: fmt.Sprintf("procurement:%s:%s", p.InvoiceNumber, pu.DrugCode),
				VmedisCode: pu.DrugCode,
			})
		}
	}

	if err := s.drugProducer.ProduceUpdatedDrugByVmedisCode(ctx, updatedDrugs); err != nil {
		return fmt.Errorf("produce updated drug by vmedis code: %w", err)
	}

	log.Printf("Produced %d updated drugs by vmedis code", len(updatedDrugs))

	return nil
}

func (s *Service) DumpRecommendationsFromVmedisToRedis(ctx context.Context) error {
	log.Println("Generating procurement recommendations and writing them to cache")

	recommendations, err := s.GenerateRecommendations(ctx)
	if err != nil {
		return fmt.Errorf("generate procurement recommendations: %w", err)
	}

	log.Printf("Writing %d procurement recommendations to Redis", len(recommendations.Recommendations))

	if err := s.redisDB.SetRecommendations(ctx, recommendations); err != nil {
		return fmt.Errorf("write procurement recommendations to Redis: %w", err)
	}

	log.Printf("Wrote %d procurement recommendations to Redis", len(recommendations.Recommendations))
	return nil
}

func (s *Service) GenerateRecommendations(ctx context.Context) (RecommendationsResponse, error) {
	log.Printf("Getting all out-of-stock-drugs for writing procurement recommendations")
	oosDrugs, err := s.vmedis.GetAllOutOfStockDrugs(ctx)
	if err != nil {
		return RecommendationsResponse{}, fmt.Errorf("get all out of stock drugs for writing procurement recommendations: %w", err)
	}

	log.Printf("Got %d out-of-stock drugs for writing procurement recommendations", len(oosDrugs))

	log.Printf("Getting drug units of out-of-stock drugs")
	drugCodes := make([]string, len(oosDrugs))
	for i, d := range oosDrugs {
		drugCodes[i] = d.Drug.VmedisCode
	}

	drugUnitsByDrugCode, err := s.drugUnitsGetter.GetDrugUnitsByDrugVmedisCodes(ctx, drugCodes)
	if err != nil {
		return RecommendationsResponse{}, fmt.Errorf("get drug units of out-of-stock drugs: %w", err)
	}

	var unitCount int
	for _, units := range drugUnitsByDrugCode {
		unitCount += len(units)
	}
	log.Printf("Got %d drug units of out-of-stock drugs", unitCount)

	recommendations := make([]Recommendation, len(oosDrugs))
	for i, drugStock := range oosDrugs {
		procurement, alternatives := calculateRecommendation(drugStock, drugUnitsByDrugCode[drugStock.Drug.VmedisCode])

		recommendations[i] = Recommendation{
			DrugStock:    drug.FromVmedisDrugStock(drugStock),
			FromSupplier: drugStock.Drug.Supplier,
			Procurement:  procurement,
			Alternatives: alternatives,
		}
	}

	return RecommendationsResponse{
		Recommendations: recommendations,
		ComputedAt:      time.Now(),
	}, nil
}

func calculateRecommendation(stock vmedis.DrugStock, drugUnits []drug.Unit) (chosen drug.Stock, alternatives []drug.Stock) {
	smallestQ := stock.Drug.MinimumStock.Quantity*2 - stock.Stock.Quantity
	if smallestQ < 2 {
		smallestQ = 2
	}

	fallback := drug.Stock{
		Unit:     stock.Stock.Unit,
		Quantity: smallestQ,
	}

	if len(drugUnits) == 0 {
		return fallback, nil
	}

	qPerUnit := make([]float64, len(drugUnits))
	qPerUnit[0] = smallestQ

	for i := 1; i < len(drugUnits); i++ {
		qPerUnit[i] = math.Round(qPerUnit[i-1] / math.Max(drugUnits[i].ConversionToParentUnit, 1))
	}

	foundChosen := false
	for i := len(drugUnits) - 1; i >= 0; i-- {
		proc := drug.Stock{
			Unit:     drugUnits[i].Unit,
			Quantity: qPerUnit[i],
		}

		if proc.Quantity > 0 {
			if !foundChosen {
				foundChosen = true
				chosen = proc
			} else {
				alternatives = append(alternatives, proc)
			}
		}
	}

	if !foundChosen {
		chosen = fallback
	}

	return
}

func (s *Service) GetInvoiceCalculators(ctx context.Context) ([]InvoiceCalculator, error) {
	calculators, err := s.db.GetInvoiceCalculators(ctx)
	if err != nil {
		return nil, fmt.Errorf("get invoice calculators from DB: %w", err)
	}

	return calculators, nil
}

func NewService(
	db *gorm.DB,
	redisClient *redis.Client,
	vmedisClient *vmedis.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) *Service {
	return &Service{
		db:              NewDatabase(db),
		redisDB:         NewRedisDatabase(redisClient),
		vmedis:          vmedisClient,
		drugProducer:    drugProducer,
		drugUnitsGetter: drugUnitsGetter,
	}
}
