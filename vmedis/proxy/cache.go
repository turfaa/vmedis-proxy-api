package proxy

import (
	"fmt"
	"time"

	"github.com/chenyahui/gin-cache/persist"
)

// CompressedCache wraps a CacheStore and compresses the data before storing it.
type CompressedCache struct {
	Store persist.CacheStore
}

// Set stores an item in the underlying CacheStore after compressing it.
func (c CompressedCache) Set(key string, value interface{}, expire time.Duration) error {
	payload, err := persist.Serialize(value)
	if err != nil {
		return fmt.Errorf("serialize: %w", err)
	}

	compressed, err := zlibCompress(payload)
	if err != nil {
		return fmt.Errorf("zlib compress: %w", err)
	}

	return c.Store.Set(key+".zlib", compressed, expire)
}

// Get retrieves an item from the underlying CacheStore and decompresses it.
func (c CompressedCache) Get(key string, value interface{}) error {
	var compressed []byte
	if err := c.Store.Get(key+".zlib", &compressed); err != nil {
		return fmt.Errorf("get: %w", err)
	}

	uncompressed, err := zlibDecompress(compressed)
	if err != nil {
		return fmt.Errorf("zlib decompress: %w", err)
	}

	return persist.Deserialize(uncompressed, value)
}

// Delete removes an item from the underlying CacheStore.
func (c CompressedCache) Delete(key string) error {
	return c.Delete(key)
}
