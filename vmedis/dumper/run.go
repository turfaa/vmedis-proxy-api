package dumper

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

const (
	// DailySalesStatisticsSchedule is the schedule of the daily sales statistics dumper.
	// It currently runs every the first second of every hour.
	DailySalesStatisticsSchedule = "0 0 * * * *"

	// DrugInterval is the interval of the drugs' dumper.
	// It currently runs every 6 hour.
	DrugInterval = 6 * time.Hour

	// ProcurementRecommendationsInterval is the interval of the procurement recommendations' dumper.
	// It currently runs every hour.
	ProcurementRecommendationsInterval = time.Hour
)

// Run runs the data dumper.
// All the dumper intervals and schedules are currently hardcoded.
func Run(vmedisClient *client.Client, db *gorm.DB, redisClient *redis.Client) {
	scheduler := gocron.NewScheduler(time.Local)

	if _, err := scheduler.CronWithSeconds(DailySalesStatisticsSchedule).Do(DumpDailySalesStatistics, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales statistics dumper: %s\n", err)
	}

	if _, err := scheduler.Every(DrugInterval).Do(DumpDrugs, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling drugs dumper: %s\n", err)
	}

	if _, err := scheduler.Every(ProcurementRecommendationsInterval).Do(DumpProcurementRecommendations, redisClient, vmedisClient); err != nil {
		log.Fatalf("Error scheduling procurement recommendations dumper: %s\n", err)
	}

	log.Println("Starting data dumper")
	scheduler.StartAsync()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	<-done

	log.Println("Stopping data dumper")
	scheduler.Stop()
}
