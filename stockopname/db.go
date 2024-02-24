package stockopname

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Database struct {
	db *gorm.DB
}

func (d *Database) GetStockOpnamesBetweenTime(ctx context.Context, from, to time.Time) ([]StockOpname, error) {
	var soModels []models.StockOpname
	if err := d.dbCtx(ctx).
		Where("date BETWEEN ? AND ?", from, to).
		Find(&soModels).
		Error; err != nil {
		return nil, fmt.Errorf("get stock opnames from DB: %w", err)
	}

	sos := make([]StockOpname, len(soModels))
	for i, so := range soModels {
		sos[i] = FromModelsStockOpname(so)
	}

	return sos, nil
}

func (d *Database) UpsertVmedisStockOpnames(ctx context.Context, stockOpnames []vmedis.StockOpname) error {
	soModels := make([]models.StockOpname, len(stockOpnames))
	for i, so := range stockOpnames {
		id := so.ID
		if id == "" {
			id = fmt.Sprintf("%s-%s-%s-%s-%d", so.DrugCode, so.BatchCode, so.Unit, so.Date.Time.Format("2006-01-02"), rand.Int())
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

	if err := d.dbCtx(ctx).
		Clauses(
			clause.OnConflict{
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
			},
		).
		Create(&soModels).
		Error; err != nil {
		return fmt.Errorf("upsert vmedis stock opnames to db: %w", err)
	}

	return nil
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}
