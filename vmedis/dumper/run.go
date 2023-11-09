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
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

const (
	// DailySalesStatisticsSchedule is the schedule of the daily sales statistics dumper.
	// It currently runs every hour + 30 seconds.
	// We add the 30 seconds delay to allow the vmedis server's cache to clear, especially in the midnight.
	DailySalesStatisticsSchedule = "30 0 * * * *"

	// DailySalesSchedule is the schedule of the daily sales dumper.
	// It currently runs every the first second of every hour.
	DailySalesSchedule = "0 * * * *"

	// DailyStockOpnamesSchedule is the schedule of the daily stock opnames dumper.
	// It currently runs every xx.30 every day.
	DailyStockOpnamesSchedule = "30 * * * *"

	// DrugSchedule is the schedule of the drugs' dumper.
	// It currently runs at 12am and 2am every day.
	DrugSchedule = "0 0,2 * * *"

	// ProcurementRecommendationsSchedule is the schedule of the procurement recommendations' dumper.
	// It currently runs at 11pm, 1am, and 3am every day.
	ProcurementRecommendationsSchedule = "0 23,1,3 * * *"
)

// Run runs the data dumper.
// All the dumper intervals and schedules are currently hardcoded.
func Run(vmedisClient *client.Client, db *gorm.DB, redisClient *redis.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db = db.WithContext(ctx)
	redisClient = redisClient.WithContext(ctx)

	drugDetailsChan, closeDrugDetailsPuller := DrugDetailsPuller(ctx, db, vmedisClient)
	defer closeDrugDetailsPuller()

	scheduler := gocron.NewScheduler(time.Local)

	if _, err := scheduler.CronWithSeconds(DailySalesStatisticsSchedule).Do(DumpDailySalesStatistics, ctx, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales statistics dumper: %s\n", err)
	}

	if _, err := scheduler.Cron(DailySalesSchedule).Do(DumpDailySales, ctx, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales dumper: %s\n", err)
	}

	if _, err := scheduler.Cron(DailyStockOpnamesSchedule).Do(DumpDailyStockOpnames, ctx, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily stock opnames dumper: %s\n", err)
	}

	if _, err := scheduler.Cron(DrugSchedule).Do(DumpDrugs, ctx, db, vmedisClient, drugDetailsChan); err != nil {
		log.Fatalf("Error scheduling drugs dumper: %s\n", err)
	}

	if _, err := scheduler.Cron(ProcurementRecommendationsSchedule).Do(DumpProcurementRecommendations, ctx, db, redisClient, vmedisClient); err != nil {
		log.Fatalf("Error scheduling procurement recommendations dumper: %s\n", err)
	}

	log.Println("Starting data dumper")
	scheduler.StartAsync()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	<-done

	log.Println("Stopping data dumper")
	scheduler.Stop()
}
