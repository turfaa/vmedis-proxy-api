package drug

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

type ApiHandlerConfig struct {
	Service                    *Service
	StockOpnameLookupStartDate time.Time
}

type ConsumerConfig struct {
	DB           *gorm.DB
	RedisClient  *redis.Client
	VmedisClient *vmedisv1.Client
	KafkaWriter  *kafka.Writer

	Brokers []string

	Concurrency int
}
