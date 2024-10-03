package procurement

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/pkg2/zstd2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
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

func (d *Database) GetAggregatedProcurementsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]AggregatedProcurement, error) {
	var procurements []AggregatedProcurement
	if err := d.dbCtx(ctx).
		Raw(
			`
SELECT 
	drugs.name AS drug_name,
	procurements.amount AS quantity,
	procurements.unit AS unit
FROM
	(
		SELECT drug_code, SUM(amount) as amount, unit
		FROM 
			procurement_units JOIN procurements ON procurement_units.invoice_number = procurements.invoice_number
		WHERE
			procurements.invoice_date BETWEEN ? AND ?
		GROUP BY drug_code, unit
	) procurements
	JOIN drugs ON procurements.drug_code = drugs.vmedis_code
ORDER BY drugs.name`,
			from,
			to,
		).
		Find(&procurements).
		Error; err != nil {
		return nil, fmt.Errorf("get aggregated procurements between %s and %s from DB: %w", from, to, err)
	}

	return procurements, nil
}

func (d *Database) GetInvoiceCalculators(ctx context.Context) ([]InvoiceCalculator, error) {
	var invoiceCalculators []models.InvoiceCalculator
	if err := d.dbCtx(ctx).
		Model(&models.InvoiceCalculator{}).
		Preload("Components").
		Order("supplier").
		Find(&invoiceCalculators).
		Error; err != nil {
		return nil, fmt.Errorf("get invoice calculators from DB: %w", err)
	}

	return slices2.Map(invoiceCalculators, FromDBInvoiceCalculator), nil
}

func (d *Database) GetLastDrugProcurements(ctx context.Context, drugCode string, limit int) ([]DrugProcurement, error) {
	limit = min(limit, 20)

	var procurements []DrugProcurement
	if err := d.dbCtx(ctx).
		Raw(
			`
SELECT 
	procurement_units.created_at,
	procurement_units.drug_code,
	procurement_units.drug_name,
	procurement_units.amount,
	procurement_units.unit,
	procurement_units.total_unit_price,
	procurement_units.invoice_number,
	procurements.invoice_date,
	procurements.supplier
FROM procurement_units
JOIN procurements ON procurement_units.invoice_number = procurements.invoice_number
WHERE procurement_units.drug_code = ?
ORDER BY procurement_units.created_at DESC
LIMIT ?
			`,
			drugCode,
			limit,
		).
		Find(&procurements).
		Error; err != nil {
		return nil, fmt.Errorf("execute SQL query: %w", err)
	}

	return procurements, nil
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
