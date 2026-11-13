package stockopname

import (
	"context"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
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

// UpsertVmedisStockOpnames upserts the stock opnames, keyed by
// VmedisStockOpnameIDs.
func (d *Database) UpsertVmedisStockOpnames(ctx context.Context, stockOpnames []vmedisv2.StockOpname) error {
	if len(stockOpnames) == 0 {
		return nil
	}

	ids := VmedisStockOpnameIDs(stockOpnames)

	soModels := make([]models.StockOpname, len(stockOpnames))
	for i, so := range stockOpnames {
		soModels[i] = models.StockOpname{
			VmedisID:            ids[i],
			Date:                datatypes.Date(so.Time),
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
					"notes",
				}),
			},
		).
		CreateInBatches(&soModels, upsertBatchSize).
		Error; err != nil {
		return fmt.Errorf("upsert vmedis stock opnames to db: %w", err)
	}

	return nil
}

// upsertBatchSize keeps each upsert statement well under Postgres's limit of
// 65535 bind parameters.
const upsertBatchSize = 500

// VmedisStockOpnameIDs returns the vmedis_id of each of the stock opnames.
//
// The id is the stock opname code, which is what the v1 scraper stored, so the
// stock opnames dumped before keep their rows. Vmedis gives several counted
// batches the same code, so a code repeated within a day gets the
// "-augmented-because-duplicate-<n>" suffix the v1 scraper gave it, numbered
// in the order of stockOpnames.
func VmedisStockOpnameIDs(stockOpnames []vmedisv2.StockOpname) []string {
	ids := make([]string, len(stockOpnames))
	visitedByDate := make(map[string]map[string]bool)

	for i, so := range stockOpnames {
		id := so.Code
		if id == "" {
			id = fmt.Sprintf("sopid-%d", so.ID)
		}

		date := so.Time.Format(time.DateOnly)
		visited := visitedByDate[date]
		if visited == nil {
			visited = make(map[string]bool)
			visitedByDate[date] = visited
		}

		if visited[id] {
			for x := 2; ; x++ {
				augmented := fmt.Sprintf("%s-augmented-because-duplicate-%d", id, x)
				if !visited[augmented] {
					id = augmented
					break
				}
			}
		}

		visited[id] = true
		ids[i] = id
	}

	return ids
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}
