package drug

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	drugsKey                            = "drugs"
	drugDetailsByVmedisCodeProcessedKey = "drug-details-by-vmedis-code-processed:%s"
	drugDetailsByVmedisIDProcessedKey   = "drug-details-by-vmedis-id-processed:%s"

	processedKeysExpiry = 30 * 24 * time.Hour
)

type Cache struct {
	redis *redis.Client
}

func (c *Cache) GetDrugs(ctx context.Context) ([]Drug, error) {
	res, err := c.redis.Get(ctx, drugsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get %s in redis: %w", drugsKey, err)
	}

	var drugs []Drug
	if err := msgpack.Unmarshal([]byte(res), &drugs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s in redis: %w", drugsKey, err)
	}

	return drugs, nil
}

func (c *Cache) SetDrugs(ctx context.Context, drugs []Drug, ttl time.Duration) error {
	bytes, err := msgpack.Marshal(drugs)
	if err != nil {
		return fmt.Errorf("failed to marshal drugs: %w", err)
	}

	if err := c.redis.Set(ctx, drugsKey, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set %s in redis: %w", drugsKey, err)
	}

	return nil
}

func (c *Cache) HasDrugDetailsByVmedisCodeProcessed(ctx context.Context, requestKey string) (bool, error) {
	redisKey := fmt.Sprintf(drugDetailsByVmedisCodeProcessedKey, requestKey)
	res, err := c.redis.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if %s exists in redis: %w", redisKey, err)
	}

	return res > 0, nil
}

func (c *Cache) MarkDrugDetailsByVmedisCodeProcessed(ctx context.Context, requestKey string) error {
	redisKey := fmt.Sprintf(drugDetailsByVmedisCodeProcessedKey, requestKey)
	if err := c.redis.SetEX(ctx, redisKey, time.Now(), processedKeysExpiry).Err(); err != nil {
		return fmt.Errorf("failed to set %s in redis: %w", redisKey, err)
	}

	return nil
}

func (c *Cache) HasDrugDetailsByVmedisIDProcessed(ctx context.Context, requestKey string) (bool, error) {
	redisKey := fmt.Sprintf(drugDetailsByVmedisIDProcessedKey, requestKey)
	res, err := c.redis.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if %s exists in redis: %w", redisKey, err)
	}

	return res > 0, nil
}

func (c *Cache) MarkDrugDetailsByVmedisIDProcessed(ctx context.Context, requestKey string) error {
	redisKey := fmt.Sprintf(drugDetailsByVmedisIDProcessedKey, requestKey)
	if err := c.redis.SetEX(ctx, redisKey, time.Now(), processedKeysExpiry).Err(); err != nil {
		return fmt.Errorf("failed to set %s in redis: %w", redisKey, err)
	}

	return nil
}

func NewCache(redisClient *redis.Client) *Cache {
	return &Cache{
		redis: redisClient,
	}
}
