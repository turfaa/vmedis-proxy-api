package cmd

import (
	"crypto/tls"
	"log"
	"net"
	"net/smtp"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/auth"
	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/pkg2/email2"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/sale"
	"github.com/turfaa/vmedis-proxy-api/stockopname"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"github.com/turfaa/vmedis-proxy-api/vmedis/token"
)

var (
	db                 atomic.Pointer[gorm.DB]
	vmedisClient       atomic.Pointer[vmedis.Client]
	redisClient        atomic.Pointer[redis.Client]
	drugProducer       atomic.Pointer[drug.Producer]
	kafkaWriter        atomic.Pointer[kafka.Writer]
	tokenProvider      atomic.Pointer[token.Provider]
	vmedisRateLimiter  atomic.Pointer[rate.Limiter]
	tokenRefresher     atomic.Pointer[token.Refresher]
	drugService        atomic.Pointer[drug.Service]
	drugDatabase       atomic.Pointer[drug.Database]
	procurementService atomic.Pointer[procurement.Service]
	saleService        atomic.Pointer[sale.Service]
	emailer            atomic.Pointer[email2.Emailer]
	stockOpnameService atomic.Pointer[stockopname.Service]
	drugHandler        atomic.Pointer[drug.ApiHandler]
	saleHandler        atomic.Pointer[sale.ApiHandler]
	procurementHandler atomic.Pointer[procurement.ApiHandler]
	stockOpnameHandler atomic.Pointer[stockopname.ApiHandler]
	authService        atomic.Pointer[auth.Service]
	authHandler        atomic.Pointer[auth.ApiHandler]
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
		getTokenProvider(),
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

func getTokenProvider() *token.Provider {
	if val := tokenProvider.Load(); val != nil {
		return val
	}

	newProvider, err := token.NewProvider(getDatabase(), viper.GetDuration("refresh_interval"))
	if err != nil {
		log.Fatalf("Error creating token provider: %s", err)
	}

	if !tokenProvider.CompareAndSwap(nil, newProvider) {
		return tokenProvider.Load()
	}

	return newProvider
}

func getTokenRefresher() *token.Refresher {
	if val := tokenRefresher.Load(); val != nil {
		return val
	}

	newRefresher := token.NewRefresher(getDatabase(), getVmedisClient())

	if !tokenRefresher.CompareAndSwap(nil, newRefresher) {
		return tokenRefresher.Load()
	}

	return newRefresher
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

func getDrugService() *drug.Service {
	if val := drugService.Load(); val != nil {
		return val
	}

	newService := drug.NewService(
		getRedisClient(),
		getDatabase(),
		getVmedisClient(),
		getKafkaWriter(),
	)

	if !drugService.CompareAndSwap(nil, newService) {
		return drugService.Load()
	}

	return newService
}

func getDrugDatabase() *drug.Database {
	if val := drugDatabase.Load(); val != nil {
		return val
	}

	newDatabase := drug.NewDatabase(getDatabase())

	if !drugDatabase.CompareAndSwap(nil, newDatabase) {
		return drugDatabase.Load()
	}

	return newDatabase
}

func getProcurementService() *procurement.Service {
	if val := procurementService.Load(); val != nil {
		return val
	}

	newService := procurement.NewService(
		getDatabase(),
		getRedisClient(),
		getVmedisClient(),
		getDrugProducer(),
		getDrugDatabase(),
	)

	if !procurementService.CompareAndSwap(nil, newService) {
		return procurementService.Load()
	}

	return newService
}

func getSaleService() *sale.Service {
	if val := saleService.Load(); val != nil {
		return val
	}

	newService := sale.NewService(
		getDatabase(),
		getVmedisClient(),
		getDrugService(),
		getDrugProducer(),
	)

	if !saleService.CompareAndSwap(nil, newService) {
		return saleService.Load()
	}

	return newService
}

func getEmailer() *email2.Emailer {
	if val := emailer.Load(); val != nil {
		return val
	}

	smtpAddress := viper.GetString("email.smtp_address")
	smtpHost, _, err := net.SplitHostPort(smtpAddress)
	if err != nil {
		log.Fatalf("Error parsing SMTP address '%s': %s", smtpAddress, err)
	}

	newPool := email2.NewEmailer(
		smtpAddress,
		smtp.PlainAuth(
			"",
			viper.GetString("email.smtp_username"),
			viper.GetString("email.smtp_password"),
			smtpHost,
		),
		&tls.Config{ServerName: smtpHost, InsecureSkipVerify: true},
	)

	if !emailer.CompareAndSwap(nil, newPool) {
		return emailer.Load()
	}

	return newPool
}

func getStockOpnameService() *stockopname.Service {
	if val := stockOpnameService.Load(); val != nil {
		return val
	}

	newService := stockopname.NewService(
		getDatabase(),
		getVmedisClient(),
		getDrugProducer(),
	)

	if !stockOpnameService.CompareAndSwap(nil, newService) {
		return stockOpnameService.Load()
	}

	return newService
}

func getDrugHandler(stockOpnameLookupStartDate time.Time) *drug.ApiHandler {
	if val := drugHandler.Load(); val != nil {
		return val
	}

	newHandler := drug.NewApiHandler(
		drug.ApiHandlerConfig{
			RedisClient:                getRedisClient(),
			Service:                    getDrugService(),
			StockOpnameLookupStartDate: stockOpnameLookupStartDate.Local(),
		},
	)

	if !drugHandler.CompareAndSwap(nil, newHandler) {
		return drugHandler.Load()
	}

	return newHandler
}

func getSaleHandler() *sale.ApiHandler {
	if val := saleHandler.Load(); val != nil {
		return val
	}

	newHandler := sale.NewApiHandler(getSaleService())

	if !saleHandler.CompareAndSwap(nil, newHandler) {
		return saleHandler.Load()
	}

	return newHandler
}

func getProcurementHandler() *procurement.ApiHandler {
	if val := procurementHandler.Load(); val != nil {
		return val
	}

	newHandler := procurement.NewApiHandler(getProcurementService())

	if !procurementHandler.CompareAndSwap(nil, newHandler) {
		return procurementHandler.Load()
	}

	return newHandler
}

func getStockOpnameHandler() *stockopname.ApiHandler {
	if val := stockOpnameHandler.Load(); val != nil {
		return val
	}

	newHandler := stockopname.NewApiHandler(getStockOpnameService())

	if !stockOpnameHandler.CompareAndSwap(nil, newHandler) {
		return stockOpnameHandler.Load()
	}

	return newHandler
}

func getAuthService() *auth.Service {
	if val := authService.Load(); val != nil {
		return val
	}

	newService := auth.NewService(getDatabase())

	if !authService.CompareAndSwap(nil, newService) {
		return authService.Load()
	}

	return newService
}

func getAuthHandler() *auth.ApiHandler {
	if val := authHandler.Load(); val != nil {
		return val
	}

	newHandler := auth.NewApiHandler(getAuthService())

	if !authHandler.CompareAndSwap(nil, newHandler) {
		return authHandler.Load()
	}

	return newHandler
}
