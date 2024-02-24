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
	"github.com/turfaa/vmedis-proxy-api/stockopname"
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
	// It currently runs at 12.25am and 2.25am every day.
	// The additional 25 minutes is due to the vmedis server being consistently down
	// at exactly 12.00am.
	DrugSchedule = "25 0,2 * * *"
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
	drugService := drug.NewService(db, vmedisClient, kafkaWriter)

	stockOpnameService := stockopname.NewService(db, vmedisClient, drugProducer)

	if _, err := scheduler.CronWithSeconds(DailySalesStatisticsSchedule).Do(DumpDailySalesStatistics, ctx, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales statistics dumper: %s", err)
	}

	if _, err := scheduler.Cron(DailySalesSchedule).Do(DumpDailySales, ctx, db, vmedisClient, drugProducer); err != nil {
		log.Fatalf("Error scheduling daily sales dumper: %s", err)
	}

	if _, err := scheduler.Cron(DailyStockOpnamesSchedule).Do(DumpDailyStockOpnames, ctx, stockOpnameService); err != nil {
		log.Fatalf("Error scheduling daily stock opnames dumper: %s", err)
	}

	if _, err := scheduler.Cron(DrugSchedule).Do(DumpDrugs, ctx, drugService); err != nil {
		log.Fatalf("Error scheduling drugs dumper: %s", err)
	}

	log.Println("Starting data dumper")
	scheduler.StartAsync()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	<-done

	log.Println("Stopping data dumper")
	scheduler.Stop()
}
