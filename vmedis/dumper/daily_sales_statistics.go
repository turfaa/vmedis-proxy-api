package dumper

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// DumpDailySalesStatistics dumps the daily sales statistics.
func DumpDailySalesStatistics(ctx context.Context, db *gorm.DB, vmedisClient *client.Client) {
	log.Println("Dumping daily sales statistics")

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	data, err := vmedisClient.GetDailySalesStatistics(ctx)
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
