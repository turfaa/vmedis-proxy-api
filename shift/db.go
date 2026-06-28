package shift

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	shiftDumpLockRedisKey = "shift:dump:lock"

	// dumpLockTTL is the maximum duration the shift dump lock is held before it
	// is automatically released by Redis.
	dumpLockTTL = 30 * time.Second
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

func (d *Database) GetShiftByCode(ctx context.Context, code string) (models.Shift, error) {
	var shift models.Shift

	if err := d.dbCtx(ctx).Where("code = ?", code).Order("started_at DESC").First(&shift).Error; err != nil {
		return models.Shift{}, fmt.Errorf("failed to get shift by code %s: %w", code, err)
	}

	return shift, nil
}

func (d *Database) GetShiftByVmedisID(ctx context.Context, vmedisID int) (models.Shift, error) {
	var shift models.Shift

	if err := d.dbCtx(ctx).Where("vmedis_id = ?", vmedisID).First(&shift).Error; err != nil {
		return models.Shift{}, fmt.Errorf("failed to get shift by vmedis id %d: %w", vmedisID, err)
	}

	return shift, nil
}

func (d *Database) GetShiftsBetween(ctx context.Context, from time.Time, to time.Time) ([]models.Shift, error) {
	var shifts []models.Shift

	if err := d.dbCtx(ctx).Where("started_at >= ? AND ended_at <= ?", from, to).Find(&shifts).Error; err != nil {
		return nil, fmt.Errorf("failed to get shifts between %s and %s from db: %w", from, to, err)
	}

	return shifts, nil
}

func (d *Database) UpsertVmedisShifts(ctx context.Context, shifts []vmedisv1.Shift) error {
	if len(shifts) == 0 {
		return nil
	}

	dbShifts := slices2.Map(shifts, vmedisShiftToDBShift)

	if err := d.dbCtx(ctx).
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "vmedis_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"updated_at",
					"code",
					"cashier",
					"started_at",
					"ended_at",
					"initial_cash",
					"expected_final_cash",
					"actual_final_cash",
					"final_cash_difference",
					"supervisor",
					"notes",
				}),
			},
		).
		Create(&dbShifts).
		Error; err != nil {
		return fmt.Errorf("failed to upsert shifts: %w", err)
	}

	return nil
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

// AcquireDumpLock attempts to acquire the shift dump lock. It returns the lock
// token and true if the lock was acquired, or an empty token and false if the
// lock is already held by another process. The returned token must be passed to
// ReleaseDumpLock to release the lock.
func (d *RedisDatabase) AcquireDumpLock(ctx context.Context) (token string, acquired bool, err error) {
	token = uuid.NewString()

	acquired, err = d.redis.SetNX(ctx, shiftDumpLockRedisKey, token, dumpLockTTL).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire shift dump lock: %w", err)
	}

	if !acquired {
		return "", false, nil
	}

	return token, true, nil
}

// ReleaseDumpLock releases the shift dump lock only if it is still held by the
// caller identified by token. Releasing a lock that has already been auto-expired
// and re-acquired by another process is a no-op.
func (d *RedisDatabase) ReleaseDumpLock(ctx context.Context, token string) error {
	if err := releaseDumpLockScript.Run(ctx, d.redis, []string{shiftDumpLockRedisKey}, token).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("release shift dump lock: %w", err)
	}

	return nil
}

// IsDumpLocked reports whether the shift dump lock is currently held by any
// process.
func (d *RedisDatabase) IsDumpLocked(ctx context.Context) (bool, error) {
	exists, err := d.redis.Exists(ctx, shiftDumpLockRedisKey).Result()
	if err != nil {
		return false, fmt.Errorf("check shift dump lock: %w", err)
	}

	return exists > 0, nil
}
