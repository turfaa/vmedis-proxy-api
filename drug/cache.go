package drug

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	drugDetailsByVmedisCodeProcessedKey = "drug-details-by-vmedis-code-processed:%s"
	drugDetailsByVmedisIDProcessedKey   = "drug-details-by-vmedis-id-processed:%s"

	processedKeysExpiry = 24 * time.Hour
)

type Cache struct {
	redis *redis.Client
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
