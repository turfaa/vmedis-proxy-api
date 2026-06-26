package procurement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/pkg2/zstd2"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

var (
	procurementRecommendationsRedisKey     = "procurement:recommendations"
	procurementRecommendationsLockRedisKey = "procurement:recommendations:lock"
)

const (
	// recommendationsLockTTL is the maximum duration the procurement recommendations
	// lock is held before it is automatically released by Redis.
	recommendationsLockTTL = 10 * time.Minute
)

// releaseRecommendationsLockScript releases the lock only if it is still held by
// the caller, identified by the token in ARGV[1]. This prevents a process from
// releasing a lock that was auto-expired and then re-acquired by another process.
var releaseRecommendationsLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`)

type Database struct {
	db *gorm.DB
}

func (d *Database) UpsertVmedisProcurements(ctx context.Context, procurements []vmedisv1.Procurement) error {
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
					"input_date",
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
	redis                         redis.UniversalClient
	shouldCompressRecommendations bool
}

func (d *RedisDatabase) GetRecommendations(ctx context.Context) (RecommendationsResponse, error) {
	raw, err := d.redis.Get(ctx, d.recommendationsKey()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return RecommendationsResponse{}, nil
		}

		return RecommendationsResponse{}, fmt.Errorf("get procurement recommendations from Redis: %w", err)
	}

	var data []byte
	if d.shouldCompressRecommendations {
		data, err = zstd2.Decompress([]byte(raw))
		if err != nil {
			return RecommendationsResponse{}, fmt.Errorf("decompress procurement recommendations: %w", err)
		}
	} else {
		data = []byte(raw)
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

	if d.shouldCompressRecommendations {
		data, err = zstd2.Compress(data)
		if err != nil {
			return fmt.Errorf("compress procurement recommendations: %w", err)
		}
	}

	return d.redis.Set(ctx, d.recommendationsKey(), data, 30*24*time.Hour).Err()
}

// AcquireRecommendationsLock attempts to acquire the procurement recommendations
// lock. It returns the lock token and true if the lock was acquired, or an empty
// token and false if the lock is already held by another process. The returned
// token must be passed to ReleaseRecommendationsLock to release the lock.
func (d *RedisDatabase) AcquireRecommendationsLock(ctx context.Context) (token string, acquired bool, err error) {
	token = uuid.NewString()

	acquired, err = d.redis.SetNX(ctx, procurementRecommendationsLockRedisKey, token, recommendationsLockTTL).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire procurement recommendations lock: %w", err)
	}

	if !acquired {
		return "", false, nil
	}

	return token, true, nil
}

// ReleaseRecommendationsLock releases the procurement recommendations lock only
// if it is still held by the caller identified by token. Releasing a lock that
// has already been auto-expired and re-acquired by another process is a no-op.
func (d *RedisDatabase) ReleaseRecommendationsLock(ctx context.Context, token string) error {
	if err := releaseRecommendationsLockScript.Run(ctx, d.redis, []string{procurementRecommendationsLockRedisKey}, token).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("release procurement recommendations lock: %w", err)
	}

	return nil
}

// IsRecommendationsLocked reports whether the procurement recommendations lock is
// currently held by any process.
func (d *RedisDatabase) IsRecommendationsLocked(ctx context.Context) (bool, error) {
	exists, err := d.redis.Exists(ctx, procurementRecommendationsLockRedisKey).Result()
	if err != nil {
		return false, fmt.Errorf("check procurement recommendations lock: %w", err)
	}

	return exists > 0, nil
}

func (d *RedisDatabase) recommendationsKey() string {
	if d.shouldCompressRecommendations {
		return procurementRecommendationsRedisKey + ".zstd"
	}

	return procurementRecommendationsRedisKey
}

func NewRedisDatabase(redisClient redis.UniversalClient) *RedisDatabase {
	return &RedisDatabase{
		redis: redisClient,
	}
}
