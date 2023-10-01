package dumper

import (
	"context"
	"log"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// DumpDailyStockOpnames dumps the daily stock opnames.
func DumpDailyStockOpnames(ctx context.Context, db *gorm.DB, vmedisClient *client.Client) {
	log.Println("Dumping stock opnames")

	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	sos, err := vmedisClient.GetAllTodayStockOpnames(ctx)
	if err != nil {
		log.Printf("Error getting stock opnames: %s\n", err)
		return
	}

	log.Printf("Got %d stock opnames\n", len(sos))

	soModels := make([]models.StockOpname, len(sos))
	for i, so := range sos {
		soModels[i] = models.StockOpname{
			VmedisID:            so.ID,
			Date:                datatypes.Date(so.Date.Time),
			DrugCode:            so.DrugCode,
			DrugName:            so.DrugName,
			BatchCode:           so.BatchCode,
			Unit:                so.Unit,
			InitialQuantity:     so.InitialQuantity,
			RealQuantity:        so.RealQuantity,
			QuantityDifference:  so.QuantityDifference,
			HPPDifference:       so.HPPDifference,
			SalePriceDifference: so.SalePriceDifference,
		}
	}

	if err := dumpStockOpnames(db, soModels); err != nil {
		log.Printf("Error dumping stock opnames: %s\n", err)
		return
	}
}

func dumpStockOpnames(db *gorm.DB, stockOpnames []models.StockOpname) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "vmedis_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"updated_at",
			"date",
			"drug_code",
			"drug_name",
			"batch_code",
			"unit",
			"initial_quantity",
			"real_quantity",
			"quantity_difference",
			"hpp_difference",
			"sale_price_difference",
		}),
	}).Create(&stockOpnames).Error
}
