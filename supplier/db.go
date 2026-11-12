package supplier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

const (
	supplierDumpLockRedisKey = "supplier:dump:lock"

	// dumpLockTTL is the maximum duration the supplier dump lock is held before
	// it is automatically released by Redis. A dump is cancelled once it runs
	// this long, so the lock never expires while a dump is still running.
	dumpLockTTL = 5 * time.Minute
)

// releaseDumpLockScript releases the lock only if it is still held by the
// caller, identified by the token in ARGV[1]. This prevents a process from
// releasing a lock that was auto-expired and then re-acquired by another process.
var releaseDumpLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

// GetSuppliers returns all the suppliers that are not soft-deleted, ordered by
// name.
func (d *Database) GetSuppliers(ctx context.Context) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	if err := d.dbCtx(ctx).Order("name").Order("id").Find(&suppliers).Error; err != nil {
		return nil, fmt.Errorf("get suppliers: %w", err)
	}

	return suppliers, nil
}

// UpsertVmedisSuppliers inserts the suppliers, or updates them by their
// vmedis IDs, reviving the ones that were soft-deleted.
func (d *Database) UpsertVmedisSuppliers(ctx context.Context, suppliers []vmedisv2.Supplier) error {
	if len(suppliers) == 0 {
		return nil
	}

	dbSuppliers := slices2.Map(suppliers, vmedisSupplierToDBSupplier)

	if err := d.dbCtx(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "vmedis_id"}},
			DoUpdates: database.UndeleteAndUpdateColumns([]string{
				"updated_at",
				"code",
				"name",
				"address",
				"city",
				"phone",
				"mobile_phone",
				"bank_account_number",
				"tax_number",
				"active",
				"lead_time_days",
				"rating",
			}),
		}).
		CreateInBatches(&dbSuppliers, 500).
		Error; err != nil {
		return fmt.Errorf("upsert suppliers: %w", err)
	}

	return nil
}

// SoftDeleteSuppliersNotIn soft-deletes the suppliers whose vmedis IDs are not
// in vmedisIDs, and returns how many were soft-deleted.
func (d *Database) SoftDeleteSuppliersNotIn(ctx context.Context, vmedisIDs []int) (int64, error) {
	res := d.dbCtx(ctx).Where("vmedis_id NOT IN ?", vmedisIDs).Delete(&models.Supplier{})
	if res.Error != nil {
		return 0, fmt.Errorf("soft-delete suppliers not in vmedis: %w", res.Error)
	}

	return res.RowsAffected, nil
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

type RedisDatabase struct {
	redis redis.UniversalClient
}

func NewRedisDatabase(redisClient redis.UniversalClient) *RedisDatabase {
	return &RedisDatabase{redis: redisClient}
}

// AcquireDumpLock attempts to acquire the supplier dump lock. It returns the
// lock token and true if the lock was acquired, or an empty token and false if
// the lock is already held by another process. The returned token must be
// passed to ReleaseDumpLock to release the lock.
func (d *RedisDatabase) AcquireDumpLock(ctx context.Context) (token string, acquired bool, err error) {
	token = uuid.NewString()

	acquired, err = d.redis.SetNX(ctx, supplierDumpLockRedisKey, token, dumpLockTTL).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire supplier dump lock: %w", err)
	}

	if !acquired {
		return "", false, nil
	}

	return token, true, nil
}

// ReleaseDumpLock releases the supplier dump lock only if it is still held by
// the caller identified by token.
func (d *RedisDatabase) ReleaseDumpLock(ctx context.Context, token string) error {
	if err := releaseDumpLockScript.Run(ctx, d.redis, []string{supplierDumpLockRedisKey}, token).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("release supplier dump lock: %w", err)
	}

	return nil
}

// IsDumpLocked reports whether the supplier dump lock is currently held by any
// process.
func (d *RedisDatabase) IsDumpLocked(ctx context.Context) (bool, error) {
	exists, err := d.redis.Exists(ctx, supplierDumpLockRedisKey).Result()
	if err != nil {
		return false, fmt.Errorf("check supplier dump lock: %w", err)
	}

	return exists > 0, nil
}
