package procurement

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"github.com/turfaa/vmedis-proxy-api/zstd2"
)

var (
	procurementRecommendationsRedisKey = "static_key.procurement_recommendations.json.zstd"
)

type Database struct {
	db *gorm.DB
}

func (d *Database) UpsertVmedisProcurements(ctx context.Context, procurements []vmedis.Procurement) error {
	if len(procurements) == 0 {
		return nil
	}

	dbProcurements := slices2.Map(procurements, vmedisProcurementToDBProcurement)

	return d.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "invoice_number"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"updated_at",
					"invoice_date",
					"supplier",
					"warehouse",
					"payment_type",
					"operator",
					"cash_discount_percentage",
					"discount_percentage",
					"discount_amount",
					"tax_percentage",
					"tax_amount",
					"miscellaneous_cost",
					"total",
				}),
			},
		).
			Omit("ProcurementUnits").
			Create(&dbProcurements).
			Error; err != nil {
			return fmt.Errorf("upsert procurements: %w", err)
		}

		for _, p := range dbProcurements {
			if len(p.ProcurementUnits) == 0 {
				log.Printf("Procurement %s has no procurement units", p.InvoiceNumber)
				continue
			}

			if err := tx.Clauses(
				clause.OnConflict{
					Columns: []clause.Column{{Name: "invoice_number"}, {Name: "id_in_procurement"}},
					DoUpdates: clause.AssignmentColumns([]string{
						"updated_at",
						"drug_code",
						"drug_name",
						"amount",
						"unit",
						"unit_base_price",
						"discount_percentage",
						"discount_two_percentage",
						"discount_three_percentage",
						"total_unit_price",
						"unit_taxed_price",
						"expiry_date",
						"batch_number",
						"total",
					}),
				},
			).
				Create(&p.ProcurementUnits).
				Error; err != nil {
				return fmt.Errorf("upsert procurement units: %w", err)
			}
		}

		return nil
	})
}

func vmedisProcurementToDBProcurement(p vmedis.Procurement) models.Procurement {
	return models.Procurement{
		InvoiceNumber:          p.InvoiceNumber,
		InvoiceDate:            datatypes.Date(p.Date.Time),
		Supplier:               p.Supplier,
		Warehouse:              p.Warehouse,
		PaymentType:            p.PaymentType,
		Operator:               p.Operator,
		CashDiscountPercentage: p.CashDiscountPercentage.Value,
		DiscountPercentage:     p.DiscountPercentage.Value,
		DiscountAmount:         p.DiscountAmount,
		TaxPercentage:          p.TaxPercentage.Value,
		TaxAmount:              p.TaxAmount,
		MiscellaneousCost:      p.MiscellaneousCost,
		Total:                  p.Total,
		ProcurementUnits: slices2.Map(p.ProcurementUnits, func(u vmedis.ProcurementUnit) models.ProcurementUnit {
			unit := vmedisProcurementUnitToDBProcurementUnit(u)
			unit.InvoiceNumber = p.InvoiceNumber
			return unit
		}),
	}
}

func vmedisProcurementUnitToDBProcurementUnit(u vmedis.ProcurementUnit) models.ProcurementUnit {
	return models.ProcurementUnit{
		IDInProcurement:         u.IDInProcurement,
		DrugCode:                u.DrugCode,
		DrugName:                u.DrugName,
		Amount:                  u.Amount,
		Unit:                    u.Unit,
		UnitBasePrice:           u.UnitBasePrice,
		DiscountPercentage:      u.DiscountPercentage.Value,
		DiscountTwoPercentage:   u.DiscountTwoPercentage.Value,
		DiscountThreePercentage: u.DiscountThreePercentage.Value,
		TotalUnitPrice:          u.TotalUnitPrice,
		UnitTaxedPrice:          u.UnitTaxedPrice,
		ExpiryDate:              u.ExpiryDate.Time,
		BatchNumber:             u.BatchNumber,
		Total:                   u.Total,
	}
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}

type RedisDatabase struct {
	redis *redis.Client
}

func (d *RedisDatabase) GetRecommendations(ctx context.Context) (RecommendationsResponse, error) {
	compressed, err := d.redis.Get(ctx, procurementRecommendationsRedisKey).Result()
	if err != nil {
		return RecommendationsResponse{}, fmt.Errorf("get procurement recommendations from Redis: %w", err)
	}

	data, err := zstd2.Decompress([]byte(compressed))
	if err != nil {
		return RecommendationsResponse{}, fmt.Errorf("decompress procurement recommendations: %w", err)
	}

	var response RecommendationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return RecommendationsResponse{}, fmt.Errorf("unmarshal procurement recommendations: %w", err)
	}

	return response, nil
}

func (d *RedisDatabase) SetRecommendations(ctx context.Context, recommendations RecommendationsResponse) error {
	data, err := json.Marshal(recommendations)
	if err != nil {
		return fmt.Errorf("marshal procurement recommendations: %w", err)
	}

	compressed, err := zstd2.Compress(data)
	if err != nil {
		return fmt.Errorf("compress procurement recommendations: %w", err)
	}

	return d.redis.Set(ctx, procurementRecommendationsRedisKey, string(compressed), 30*24*time.Hour).Err()
}

func NewRedisDatabase(redisClient *redis.Client) *RedisDatabase {
	return &RedisDatabase{
		redis: redisClient,
	}
}
