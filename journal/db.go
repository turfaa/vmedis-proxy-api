package journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

// Database stores the journal entry requests.
type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

// GetByKey returns the request with the idempotency key, or ErrNotFound.
func (d *Database) GetByKey(ctx context.Context, key string) (models.JournalEntryRequest, error) {
	var req models.JournalEntryRequest
	if err := d.db.WithContext(ctx).Where("idempotency_key = ?", key).Take(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.JournalEntryRequest{}, ErrNotFound
		}
		return models.JournalEntryRequest{}, fmt.Errorf("get journal entry request %q: %w", key, err)
	}

	return req, nil
}

// Insert inserts req, and reports false without inserting when its
// idempotency key is taken.
func (d *Database) Insert(ctx context.Context, req *models.JournalEntryRequest) (bool, error) {
	res := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "idempotency_key"}}, DoNothing: true}).
		Create(req)
	if res.Error != nil {
		return false, fmt.Errorf("insert journal entry request: %w", res.Error)
	}

	return res.RowsAffected == 1, nil
}

// Claim marks the request pending and counts an attempt, if no one else
// claimed it since it was read, and reports whether it did. Of two processes
// claiming the same request, only one wins.
func (d *Database) Claim(ctx context.Context, req *models.JournalEntryRequest, now time.Time) (bool, error) {
	res := d.db.WithContext(ctx).Model(&models.JournalEntryRequest{}).
		Where("id = ? AND attempts = ?", req.ID, req.Attempts).
		Updates(map[string]any{"status": models.JournalEntryRequestPending, "attempts": req.Attempts + 1, "updated_at": now})
	if res.Error != nil {
		return false, fmt.Errorf("claim journal entry request %d: %w", req.ID, res.Error)
	}

	if res.RowsAffected == 1 {
		req.Status = models.JournalEntryRequestPending
		req.Attempts++
		req.UpdatedAt = now
	}
	return res.RowsAffected == 1, nil
}

// Save saves every field of req.
func (d *Database) Save(ctx context.Context, req *models.JournalEntryRequest) error {
	if err := d.db.WithContext(ctx).Save(req).Error; err != nil {
		return fmt.Errorf("save journal entry request %d: %w", req.ID, err)
	}

	return nil
}

const (
	createLockRedisKey   = "vmedis-proxy:journal:create-lock"
	accountsRedisKey     = "vmedis-proxy:journal:accounts"
	accountsCacheTTL     = 24 * time.Hour
	createLockTTL        = 3 * time.Minute
	createLockRetryEvery = 200 * time.Millisecond
)

var releaseCreateLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

// RedisDatabase holds the create lock and the cached account list.
type RedisDatabase struct {
	redis redis.UniversalClient
}

func NewRedisDatabase(redisClient redis.UniversalClient) *RedisDatabase {
	return &RedisDatabase{redis: redisClient}
}

// AcquireCreateLock waits until it holds the lock that serialises reserving a
// No. Bukti and creating the entry with it, across every replica. It returns
// the function that releases it.
func (d *RedisDatabase) AcquireCreateLock(ctx context.Context) (func(), error) {
	token := uuid.NewString()

	for {
		acquired, err := d.redis.SetNX(ctx, createLockRedisKey, token, createLockTTL).Result()
		if err != nil {
			return nil, fmt.Errorf("acquire journal create lock: %w", err)
		}

		if acquired {
			return func() {
				// Released even when ctx is done.
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = releaseCreateLockScript.Run(ctx, d.redis, []string{createLockRedisKey}, token).Err()
			}, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for journal create lock: %w", ctx.Err())
		case <-time.After(createLockRetryEvery):
		}
	}
}

// GetAccounts returns the cached account list, or nil when it isn't cached.
func (d *RedisDatabase) GetAccounts(ctx context.Context) ([]vmedisv2.JournalAccount, error) {
	b, err := d.redis.Get(ctx, accountsRedisKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get cached journal accounts: %w", err)
	}

	var accounts []vmedisv2.JournalAccount
	if err := json.Unmarshal(b, &accounts); err != nil {
		return nil, nil
	}
	return accounts, nil
}

// SetAccounts caches the account list for a day.
func (d *RedisDatabase) SetAccounts(ctx context.Context, accounts []vmedisv2.JournalAccount) error {
	b, err := json.Marshal(accounts)
	if err != nil {
		return fmt.Errorf("marshal journal accounts: %w", err)
	}

	if err := d.redis.Set(ctx, accountsRedisKey, b, accountsCacheTTL).Err(); err != nil {
		return fmt.Errorf("cache journal accounts: %w", err)
	}
	return nil
}
