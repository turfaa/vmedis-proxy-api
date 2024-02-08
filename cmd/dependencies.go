package cmd

import (
	"log"
	"sync/atomic"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"github.com/turfaa/vmedis-proxy-api/vmedis/token"
)

var (
	db                atomic.Pointer[gorm.DB]
	vmedisClient      atomic.Pointer[vmedis.Client]
	redisClient       atomic.Pointer[redis.Client]
	drugProducer      atomic.Pointer[drug.Producer]
	kafkaWriter       atomic.Pointer[kafka.Writer]
	tokenManager      atomic.Pointer[token.Manager]
	vmedisMiniClient  atomic.Pointer[vmedis.MiniClient]
	vmedisRateLimiter atomic.Pointer[rate.Limiter]
)

func getDatabase() *gorm.DB {
	if val := db.Load(); val != nil {
		return val
	}

	var (
		newDB *gorm.DB
		err   error
	)

	if viper.GetString("postgres_dsn") == "" {
		newDB, err = database.SqliteDB(viper.GetString("sqlite_path"))
	} else {
		newDB, err = database.PostgresDB(viper.GetString("postgres_dsn"))
	}

	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	if !db.CompareAndSwap(nil, newDB) {
		return db.Load()
	}

	return newDB
}

func getVmedisClient() *vmedis.Client {
	if val := vmedisClient.Load(); val != nil {
		return val
	}

	newClient := vmedis.New(
		viper.GetString("base_url"),
		viper.GetInt("concurrency"),
		getVmedisRateLimiter(),
		getTokenManager(),
	)

	if !vmedisClient.CompareAndSwap(nil, newClient) {
		return vmedisClient.Load()
	}

	return newClient
}

func getRedisClient() *redis.Client {
	if val := redisClient.Load(); val != nil {
		return val
	}

	newClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_address"),
		Password: viper.GetString("redis_password"),
		DB:       viper.GetInt("redis_db"),
	})

	if !redisClient.CompareAndSwap(nil, newClient) {
		return redisClient.Load()
	}

	return newClient
}

func getDrugProducer() *drug.Producer {
	if val := drugProducer.Load(); val != nil {
		return val
	}

	newProducer := drug.NewProducer(getKafkaWriter())

	if !drugProducer.CompareAndSwap(nil, newProducer) {
		return drugProducer.Load()
	}

	return newProducer
}

func getKafkaWriter() *kafka.Writer {
	if val := kafkaWriter.Load(); val != nil {
		return val
	}

	newWriter := &kafka.Writer{
		Addr:         kafka.TCP(viper.GetStringSlice("kafka_brokers")...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}

	if !kafkaWriter.CompareAndSwap(nil, newWriter) {
		return kafkaWriter.Load()
	}

	return newWriter
}

func getTokenManager() *token.Manager {
	if val := tokenManager.Load(); val != nil {
		return val
	}

	newManager, err := token.NewManager(
		getDatabase(),
		getVmedisMiniClient(),
		viper.GetDuration("refresh_interval"),
	)
	if err != nil {
		log.Fatalf("Error creating token manager: %s", err)
	}

	if !tokenManager.CompareAndSwap(nil, newManager) {
		return tokenManager.Load()
	}

	return newManager
}

func getVmedisMiniClient() *vmedis.MiniClient {
	if val := vmedisMiniClient.Load(); val != nil {
		return val
	}

	newClient := vmedis.NewMini(
		viper.GetString("base_url"),
		getVmedisRateLimiter(),
	)

	if !vmedisMiniClient.CompareAndSwap(nil, newClient) {
		return vmedisMiniClient.Load()
	}

	return newClient
}

func getVmedisRateLimiter() *rate.Limiter {
	if val := vmedisRateLimiter.Load(); val != nil {
		return val
	}

	newLimiter := rate.NewLimiter(rate.Limit(viper.GetFloat64("rate_limit")), 1)
	if !vmedisRateLimiter.CompareAndSwap(nil, newLimiter) {
		return vmedisRateLimiter.Load()
	}

	return newLimiter
}
