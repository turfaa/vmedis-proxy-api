package cmd

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/smtp"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/auth"
	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/pkg2/email2"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/purchaseorder"
	"github.com/turfaa/vmedis-proxy-api/rejecteddrug"
	"github.com/turfaa/vmedis-proxy-api/sale"
	"github.com/turfaa/vmedis-proxy-api/shift"
	"github.com/turfaa/vmedis-proxy-api/stockopname"
	"github.com/turfaa/vmedis-proxy-api/supplier"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
	token2 "github.com/turfaa/vmedis-proxy-api/vmedis/v1/token"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

var (
	db                 atomic.Pointer[gorm.DB]
	vmedisClient       atomic.Pointer[vmedisv1.Client]
	vmedisV2Client     atomic.Pointer[vmedisv2.Client]
	redisClient        atomic.Pointer[redis.UniversalClient]
	drugProducer       atomic.Pointer[drug.Producer]
	kafkaWriter        atomic.Pointer[kafka.Writer]
	tokenProvider      atomic.Pointer[token2.Provider]
	vmedisRateLimiter  atomic.Pointer[rate.Limiter]
	tokenRefresher     atomic.Pointer[token2.Refresher]
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
	shiftService       atomic.Pointer[shift.Service]
	shiftHandler       atomic.Pointer[shift.ApiHandler]
	tokenService       atomic.Pointer[token2.Service]
	tokenHandler       atomic.Pointer[token2.Handler]

	rejectedDrugService atomic.Pointer[rejecteddrug.Service]
	rejectedDrugHandler atomic.Pointer[rejecteddrug.ApiHandler]

	supplierService atomic.Pointer[supplier.Service]
	supplierHandler atomic.Pointer[supplier.ApiHandler]

	purchaseOrderService atomic.Pointer[purchaseorder.Service]
	purchaseOrderHandler atomic.Pointer[purchaseorder.ApiHandler]
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

func getVmedisClient() *vmedisv1.Client {
	if val := vmedisClient.Load(); val != nil {
		return val
	}

	newClient := vmedisv1.New(
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

func getVmedisV2Client() *vmedisv2.Client {
	if val := vmedisV2Client.Load(); val != nil {
		return val
	}

	newClient := vmedisv2.New(
		vmedisv2.NewGatewayClient(
			viper.GetString("vmedis_v2.gateway_url"),
			vmedisv2.NewCrypt(viper.GetString("vmedis_v2.crypt_key")),
		),
		viper.GetString("vmedis_v2.app_id"),
		viper.GetString("vmedis_v2.reg"),
	)

	if !vmedisV2Client.CompareAndSwap(nil, newClient) {
		return vmedisV2Client.Load()
	}

	return newClient
}

func getRedisClient() redis.UniversalClient {
	if val := redisClient.Load(); val != nil {
		return *val
	}

	options := &redis.UniversalOptions{
		Addrs:    []string{viper.GetString("redis_address")},
		Password: viper.GetString("redis_password"),
		DB:       viper.GetInt("redis_db"),
	}

	if viper.GetBool("redis_sentinel_mode") {
		options.MasterName = viper.GetString("redis_sentinel_master_name")
	}

	newClient := redis.NewUniversalClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := newClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Error pinging Redis: %s", err)
	}

	if !redisClient.CompareAndSwap(nil, &newClient) {
		return *redisClient.Load()
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

func getTokenProvider() *token2.Provider {
	if val := tokenProvider.Load(); val != nil {
		return val
	}

	// The v1 client opens v1 pages, so it can only use v1 tokens.
	newProvider, err := token2.NewProvider(getDatabase(), models.TokenVersionV1, viper.GetDuration("refresh_interval"))
	if err != nil {
		log.Fatalf("Error creating token provider: %s", err)
	}

	if !tokenProvider.CompareAndSwap(nil, newProvider) {
		return tokenProvider.Load()
	}

	return newProvider
}

func getTokenRefresher() *token2.Refresher {
	if val := tokenRefresher.Load(); val != nil {
		return val
	}

	newRefresher := token2.NewRefresher(getDatabase(), getVmedisClient())

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
		getVmedisV2Client(),
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
		getVmedisV2Client(),
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
		getVmedisV2Client(),
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
		getVmedisV2Client(),
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

	newService := auth.NewService(getRedisClient(), getDatabase())

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

func getShiftService() *shift.Service {
	if val := shiftService.Load(); val != nil {
		return val
	}

	newService := shift.NewService(getDatabase(), getRedisClient(), getVmedisV2Client())

	if !shiftService.CompareAndSwap(nil, newService) {
		return shiftService.Load()
	}

	return newService
}

func getShiftHandler() *shift.ApiHandler {
	if val := shiftHandler.Load(); val != nil {
		return val
	}

	newHandler := shift.NewApiHandler(getShiftService())

	if !shiftHandler.CompareAndSwap(nil, newHandler) {
		return shiftHandler.Load()
	}

	return newHandler
}

func getTokenService() *token2.Service {
	if val := tokenService.Load(); val != nil {
		return val
	}

	newService := token2.NewService(getDatabase(), getRedisClient(), getTokenRefresher())

	if !tokenService.CompareAndSwap(nil, newService) {
		return tokenService.Load()
	}

	return newService
}

func getTokenHandler() *token2.Handler {
	if val := tokenHandler.Load(); val != nil {
		return val
	}

	newHandler := token2.NewHandler(getTokenService())

	if !tokenHandler.CompareAndSwap(nil, newHandler) {
		return tokenHandler.Load()
	}

	return newHandler
}

func getRejectedDrugService() *rejecteddrug.Service {
	if val := rejectedDrugService.Load(); val != nil {
		return val
	}

	newService := rejecteddrug.NewService(getDatabase())

	if !rejectedDrugService.CompareAndSwap(nil, newService) {
		return rejectedDrugService.Load()
	}

	return newService
}

func getRejectedDrugHandler() *rejecteddrug.ApiHandler {
	if val := rejectedDrugHandler.Load(); val != nil {
		return val
	}

	newHandler := rejecteddrug.NewApiHandler(getRejectedDrugService())

	if !rejectedDrugHandler.CompareAndSwap(nil, newHandler) {
		return rejectedDrugHandler.Load()
	}

	return newHandler
}

func getSupplierService() *supplier.Service {
	if val := supplierService.Load(); val != nil {
		return val
	}

	newService := supplier.NewService(getDatabase(), getRedisClient(), getVmedisV2Client())

	if !supplierService.CompareAndSwap(nil, newService) {
		return supplierService.Load()
	}

	return newService
}

func getSupplierHandler() *supplier.ApiHandler {
	if val := supplierHandler.Load(); val != nil {
		return val
	}

	newHandler := supplier.NewApiHandler(getSupplierService())

	if !supplierHandler.CompareAndSwap(nil, newHandler) {
		return supplierHandler.Load()
	}

	return newHandler
}

func getPurchaseOrderConfig() purchaseorder.Config {
	return purchaseorder.Config{
		PharmacyName:            viper.GetString("pharmacy_name"),
		PharmacyLicenseNumber:   viper.GetString("pharmacy_license_number"),
		PharmacyAddress:         viper.GetString("pharmacy_address"),
		PharmacyEmail:           viper.GetString("pharmacy_email"),
		PharmacyPhone:           viper.GetString("pharmacy_phone"),
		PharmacyCity:            viper.GetString("pharmacy_city"),
		PharmacistName:          viper.GetString("pharmacist_name"),
		PharmacistLicenseNumber: viper.GetString("pharmacist_license_number"),
		VmedisBaseURL:           viper.GetString("base_url"),
	}
}

func getPurchaseOrderService() *purchaseorder.Service {
	if val := purchaseOrderService.Load(); val != nil {
		return val
	}

	config := getPurchaseOrderConfig()
	if err := config.Validate(); err != nil {
		log.Fatalf("Error configuring purchase orders: %s", err)
	}

	if !config.Enabled() {
		log.Println("Purchase orders are not configured, their endpoints respond with 503")
	}

	newService := purchaseorder.NewService(getDatabase(), config)

	if !purchaseOrderService.CompareAndSwap(nil, newService) {
		return purchaseOrderService.Load()
	}

	return newService
}

func getPurchaseOrderHandler() *purchaseorder.ApiHandler {
	if val := purchaseOrderHandler.Load(); val != nil {
		return val
	}

	newHandler := purchaseorder.NewApiHandler(getPurchaseOrderService())

	if !purchaseOrderHandler.CompareAndSwap(nil, newHandler) {
		return purchaseOrderHandler.Load()
	}

	return newHandler
}
