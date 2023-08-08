package dumper

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

const (
	// DailySalesStatisticsSchedule is the schedule of the daily sales statistics dumper.
	// It currently runs every 1 hour.
	DailySalesStatisticsSchedule = "0 0 * * * *"
)

// Run runs the data dumper.
// All the dumper intervals and schedules are currently hardcoded.
func Run(vmedisClient *client.Client) {
	db, err := database.SqliteDB("data/db.sqlite")
	if err != nil {
		log.Fatalf("Error opening database: %s\n", err)
	}

	scheduler := gocron.NewScheduler(time.Local)

	if _, err := scheduler.CronWithSeconds(DailySalesStatisticsSchedule).Do(DumpDailySalesStatistics, db, vmedisClient); err != nil {
		log.Fatalf("Error scheduling daily sales statistics dumper: %s\n", err)
	}

	log.Println("Starting data dumper")
	scheduler.StartAsync()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, os.Kill)

	<-done

	log.Println("Stopping data dumper")
	scheduler.Stop()
}

// DumpDailySalesStatistics dumps the daily sales statistics.
func DumpDailySalesStatistics(db *gorm.DB, vmedisClient *client.Client) {
	log.Println("Dumping daily sales statistics")

	data, err := vmedisClient.GetDailySalesStatistics()
	if err != nil {
		log.Printf("Error getting daily sales statistics: %s\n", err)
		return
	}

	salesFloat, err := data.TotalSalesFloat64()
	if err != nil {
		log.Printf("Error parsing total sales (%s): %s\n", data.TotalSales, err)
		return
	}

	if err := db.Create(&models.SaleStatistics{
		PulledAt:      time.Now(),
		TotalSales:    salesFloat,
		NumberOfSales: data.NumberOfSales,
	}).Error; err != nil {
		log.Printf("Error creating sale statistics: %s\n", err)
		return
	}

	log.Printf("Dumped daily sales statistics: (total sales: %f, number of sales: %d)\n", salesFloat, data.NumberOfSales)
}
