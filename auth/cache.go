package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	userKeyPrefix = "user:"
)

type Cache struct {
	redis *redis.Client
}

func NewCache(redis *redis.Client) *Cache {
	return &Cache{redis: redis}
}

func (c *Cache) GetUser(ctx context.Context, email string) (User, error) {
	res, err := c.redis.Get(ctx, userKeyPrefix+email).Result()
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	var user User
	if err := msgpack.Unmarshal([]byte(res), &user); err != nil {
		return User{}, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user, nil
}

func (c *Cache) SetUser(ctx context.Context, user User, ttl time.Duration) error {
	bytes, err := msgpack.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	redisKey := userKeyPrefix + user.Email
	if err := c.redis.Set(ctx, redisKey, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set %s in redis: %w", redisKey, err)
	}

	return nil
}
