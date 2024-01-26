package dumper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// DumpDailyStockOpnames dumps the daily stock opnames.
func DumpDailyStockOpnames(ctx context.Context, db *gorm.DB, vmedisClient *vmedis.Client) {
	log.Println("Dumping stock opnames")

	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	sos, err := vmedisClient.GetAllTodayStockOpnames(ctx)
	if err != nil {
		log.Printf("Error getting stock opnames: %s\n", err)
		return
	}

	log.Printf("Got %d stock opnames\n", len(sos))

	if len(sos) == 0 {
		log.Printf("No stock opnames found\n")
		return
	}

	soModels := make([]models.StockOpname, len(sos))
	for i, so := range sos {
		id := so.ID
		if id == "" {
			id = fmt.Sprintf("%s-%s-%s-%s-%d", so.DrugCode, so.BatchCode, so.Unit, so.Date.Time.Format("2006-01-02"), rnd.Int())
		}

		soModels[i] = models.StockOpname{
			VmedisID:            id,
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
			Notes:               so.Notes,
		}
	}

	log.Printf("Dumping %d stock opnames\n", len(soModels))
	if err := dumpStockOpnames(db, soModels); err != nil {
		log.Printf("Error dumping stock opnames: %s\n", err)
		return
	}

	log.Println("Stock opnames dumped")
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
