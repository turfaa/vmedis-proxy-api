package cmd

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

var (
	db     *gorm.DB
	dbErr  error
	dbOnce sync.Once

	vmedisClient     *vmedis.Client
	vmedisClientOnce sync.Once

	redisClient     *redis.Client
	redisClientOnce sync.Once

	drugProducer     *drug.Producer
	drugProducerOnce sync.Once

	kafkaWriter     *kafka.Writer
	kafkaWriterOnce sync.Once
)

func getDatabase() (*gorm.DB, error) {
	dbOnce.Do(func() {
		if viper.GetString("postgres_dsn") != "" {
			db, dbErr = database.PostgresDB(viper.GetString("postgres_dsn"))
			return
		}

		db, dbErr = database.SqliteDB(viper.GetString("sqlite_path"))
	})

	return db, dbErr
}

func getVmedisClient() *vmedis.Client {
	vmedisClientOnce.Do(func() {
		vmedisClient = vmedis.New(
			viper.GetString("base_url"),
			viper.GetStringSlice("session_ids"),
			viper.GetInt("concurrency"),
			rate.NewLimiter(rate.Limit(viper.GetFloat64("rate_limit")), 1),
		)
	})

	return vmedisClient
}

func getRedisClient() *redis.Client {
	redisClientOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     viper.GetString("redis_address"),
			Password: viper.GetString("redis_password"),
			DB:       viper.GetInt("redis_db"),
		})
	})

	return redisClient
}

func getDrugProducer() *drug.Producer {
	drugProducerOnce.Do(func() {
		drugProducer = drug.NewProducer(getKafkaWriter())
	})

	return drugProducer
}

func getKafkaWriter() *kafka.Writer {
	kafkaWriterOnce.Do(func() {
		kafkaWriter = &kafka.Writer{
			Addr:         kafka.TCP(viper.GetStringSlice("kafka_brokers")...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Compression:  kafka.Snappy,
		}
	})

	return kafkaWriter
}
