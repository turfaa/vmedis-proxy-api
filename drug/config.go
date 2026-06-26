package drug

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

type ApiHandlerConfig struct {
	Service                    *Service
	StockOpnameLookupStartDate time.Time
}

type ConsumerConfig struct {
	DB           *gorm.DB
	RedisClient  redis.UniversalClient
	VmedisClient *vmedisv1.Client
	KafkaWriter  *kafka.Writer

	Brokers []string

	Concurrency int
}
