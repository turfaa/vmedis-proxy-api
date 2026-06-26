package gin2

import (
	"context"
	"errors"
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/redis/go-redis/v9"
)

type RedisCacheAdapter struct {
	client redis.UniversalClient
}

func NewRedisCacheAdapter(client redis.UniversalClient) *RedisCacheAdapter {
	return &RedisCacheAdapter{client: client}
}

// Set put key value pair to redis, and expire after expireDuration
func (r *RedisCacheAdapter) Set(key string, value any, expire time.Duration) error {
	key = r.buildKey(key)

	payload, err := persist.Serialize(value)
	if err != nil {
		return err
	}

	return r.client.Set(context.Background(), key, payload, expire).Err()
}

func (r *RedisCacheAdapter) Delete(key string) error {
	key = r.buildKey(key)
	return r.client.Del(context.Background(), key).Err()
}

// Get retrieves an item from redis, if key doesn't exist, return ErrCacheMiss
func (r *RedisCacheAdapter) Get(key string, value any) error {
	key = r.buildKey(key)
	payload, err := r.client.Get(context.Background(), key).Bytes()

	if errors.Is(err, redis.Nil) {
		return persist.ErrCacheMiss
	}

	if err != nil {
		return err
	}
	return persist.Deserialize(payload, value)
}

func (r *RedisCacheAdapter) buildKey(key string) string {
	return "gin-cache:" + key
}
