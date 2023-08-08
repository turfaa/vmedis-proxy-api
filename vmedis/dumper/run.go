package dumper

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

const (
	// DailySalesStatisticsSchedule is the schedule of the daily sales statistics dumper.
	// It currently runs every the first second of every hour.
	DailySalesStatisticsSchedule = "0 0 * * * *"

	// DrugsSchedule is the schedule of the drugs' dumper.
	// It currently runs every 6 hour.
	DrugsSchedule = 6 * time.Hour
)

// Run runs the data dumper.
// All the dumper intervals and schedules are currently hardcoded.
func Run(vmedisClient *client.Client, db *gorm.DB) {
	scheduler := gocron.NewScheduler(time.Local)

	if _, err := scheduler.CronWithSeconds(DailySalesStatisticsSchedule).Do(DumpDailySalesStatistics, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales statistics dumper: %s\n", err)
	}

	if _, err := scheduler.Every(DrugsSchedule).Do(DumpDrugs, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling drugs dumper: %s\n", err)
	}

	log.Println("Starting data dumper")
	scheduler.StartAsync()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	<-done

	log.Println("Stopping data dumper")
	scheduler.Stop()
}
