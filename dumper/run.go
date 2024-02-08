package dumper

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

const (
	// DailySalesStatisticsSchedule is the schedule of the daily sales statistics dumper.
	// It currently runs every hour - 30 seconds.
	DailySalesStatisticsSchedule = "30 59 * * * *"

	// DailySalesSchedule is the schedule of the daily sales dumper.
	// It currently runs every 15 minutes.
	DailySalesSchedule = "*/15 * * * *"

	// DailyStockOpnamesSchedule is the schedule of the daily stock opnames dumper.
	// It currently runs every 10 minutes every day.
	DailyStockOpnamesSchedule = "*/10 * * * *"

	// DrugSchedule is the schedule of the drugs' dumper.
	// It currently runs at 12am and 2am every day.
	DrugSchedule = "0 0,2 * * *"

	// ProcurementRecommendationsSchedule is the schedule of the procurement recommendations' dumper.
	// It currently runs at 11pm, 1am, and 3am every day.
	ProcurementRecommendationsSchedule = "0 23,1,3 * * *"

	// ProcurementsSchedule is the schedule of the procurements' dumper.
	// It currently runs every 30 minutes at xx.05 and xx.35.
	ProcurementsSchedule = "5,35 * * * * *"
)

// Run runs the data dumper.
// All the dumper intervals and schedules are currently hardcoded.
func Run(vmedisClient *vmedis.Client, db *gorm.DB, redisClient *redis.Client, kafkaWriter *kafka.Writer) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db = db.WithContext(ctx)
	redisClient = redisClient.WithContext(ctx)

	scheduler := gocron.NewScheduler(time.Local)

	drugProducer := drug.NewProducer(kafkaWriter)
	drugDB := drug.NewDatabase(db)

	drugService := drug.NewService(db, vmedisClient, kafkaWriter)
	procurementService := procurement.NewService(db, redisClient, vmedisClient, drugProducer, drugDB)

	if _, err := scheduler.CronWithSeconds(DailySalesStatisticsSchedule).Do(DumpDailySalesStatistics, ctx, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales statistics dumper: %s", err)
	}

	if _, err := scheduler.Cron(DailySalesSchedule).Do(DumpDailySales, ctx, db, vmedisClient, drugProducer); err != nil {
		log.Fatalf("Error scheduling daily sales dumper: %s", err)
	}

	if _, err := scheduler.Cron(DailyStockOpnamesSchedule).Do(DumpDailyStockOpnames, ctx, db, vmedisClient, drugProducer); err != nil {
		log.Fatalf("Error scheduling daily stock opnames dumper: %s", err)
	}

	if _, err := scheduler.Cron(DrugSchedule).Do(DumpDrugs, ctx, drugService); err != nil {
		log.Fatalf("Error scheduling drugs dumper: %s", err)
	}

	if _, err := scheduler.Cron(ProcurementRecommendationsSchedule).Do(DumpProcurementRecommendations, ctx, procurementService); err != nil {
		log.Fatalf("Error scheduling procurement recommendations dumper: %s", err)
	}

	if _, err := scheduler.Cron(ProcurementsSchedule).Do(DumpProcurements, ctx, procurementService); err != nil {
		log.Fatalf("Error scheduling procurements dumper: %s", err)
	}

	log.Println("Starting data dumper")
	scheduler.StartAsync()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	<-done

	log.Println("Stopping data dumper")
	scheduler.Stop()
}
