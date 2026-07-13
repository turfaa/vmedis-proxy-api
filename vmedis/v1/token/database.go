package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

const (
	tokenRefreshLockRedisKey = "token:refresh:lock"

	// refreshLockTTL is the maximum duration the token refresh lock is held
	// before it is automatically released by Redis.
	refreshLockTTL = time.Minute
)

// releaseRefreshLockScript releases the lock only if it is still held by the
// caller, identified by the token in ARGV[1]. This prevents a process from
// releasing a lock that was auto-expired and then re-acquired by another process.
var releaseRefreshLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`)

type Database struct {
	db *gorm.DB
}

func (d *Database) Transaction(ctx context.Context, f func(d *Database) error) error {
	tx := d.withContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}

	if err := f(NewDatabase(tx)); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (d *Database) GetAllTokens(ctx context.Context) ([]models.VmedisToken, error) {
	var tokens []models.VmedisToken
	if err := d.withContext(ctx).Order("id ASC").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("get all tokens from DB: %w", err)
	}

	return tokens, nil
}

func (d *Database) GetNonExpiredTokens(ctx context.Context) ([]models.VmedisToken, error) {
	var tokens []models.VmedisToken
	if err := d.withContext(ctx).Where("state != 'EXPIRED'").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("get non expired tokens from DB: %w", err)
	}

	return tokens, nil
}

func (d *Database) GetActiveTokens(ctx context.Context) ([]models.VmedisToken, error) {
	var tokens []models.VmedisToken
	if err := d.withContext(ctx).Where("state = 'ACTIVE'").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("get active tokens from DB: %w", err)
	}

	return tokens, nil
}

func (d *Database) UpsertTokensState(ctx context.Context, tokens []models.VmedisToken) error {
	if len(tokens) == 0 {
		return nil
	}

	// Clear auto-assignable fields.
	for i := range tokens {
		tokens[i].ID = 0
		tokens[i].CreatedAt = time.Time{}
		tokens[i].UpdatedAt = time.Time{}
	}

	if err := d.withContext(ctx).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "token"}},
				DoUpdates: clause.AssignmentColumns([]string{"updated_at", "state"}),
			},
		).
		Create(&tokens).
		Error; err != nil {
		return fmt.Errorf("upsert tokens state: %w", err)
	}

	return nil
}

func (d *Database) InsertToken(ctx context.Context, token string) error {
	if err := d.withContext(ctx).Create(&models.VmedisToken{Token: token}).Error; err != nil {
		return fmt.Errorf("insert token: %w", err)
	}

	return nil
}

func (d *Database) DeleteToken(ctx context.Context, id uint) error {
	if err := d.withContext(ctx).Delete(&models.VmedisToken{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete token: %w", err)
	}

	return nil
}

// DeleteExpiredTokens deletes all tokens whose state is EXPIRED. It returns the
// number of tokens deleted.
func (d *Database) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	result := d.withContext(ctx).Delete(&models.VmedisToken{}, "state = 'EXPIRED'")
	if result.Error != nil {
		return 0, fmt.Errorf("delete expired tokens: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func (d *Database) withContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}

// RedisDatabase manages the distributed token refresh lock in Redis.
type RedisDatabase struct {
	redis redis.UniversalClient
}

func NewRedisDatabase(redisClient redis.UniversalClient) *RedisDatabase {
	return &RedisDatabase{redis: redisClient}
}

// AcquireRefreshLock attempts to acquire the token refresh lock. It returns the
// lock token and true if the lock was acquired, or an empty token and false if
// the lock is already held by another process. The returned token must be passed
// to ReleaseRefreshLock to release the lock.
func (d *RedisDatabase) AcquireRefreshLock(ctx context.Context) (token string, acquired bool, err error) {
	token = uuid.NewString()

	acquired, err = d.redis.SetNX(ctx, tokenRefreshLockRedisKey, token, refreshLockTTL).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire token refresh lock: %w", err)
	}

	if !acquired {
		return "", false, nil
	}

	return token, true, nil
}

// ReleaseRefreshLock releases the token refresh lock only if it is still held by
// the caller identified by token. Releasing a lock that has already been
// auto-expired and re-acquired by another process is a no-op.
func (d *RedisDatabase) ReleaseRefreshLock(ctx context.Context, token string) error {
	if err := releaseRefreshLockScript.Run(ctx, d.redis, []string{tokenRefreshLockRedisKey}, token).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("release token refresh lock: %w", err)
	}

	return nil
}

// IsRefreshLocked reports whether the token refresh lock is currently held by
// any process.
func (d *RedisDatabase) IsRefreshLocked(ctx context.Context) (bool, error) {
	exists, err := d.redis.Exists(ctx, tokenRefreshLockRedisKey).Result()
	if err != nil {
		return false, fmt.Errorf("check token refresh lock: %w", err)
	}

	return exists > 0, nil
}
