package sale

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Database struct {
	db *gorm.DB
}

func (d *Database) GetSalesBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]Sale, error) {
	var salesModels []models.Sale
	if err := d.dbCtx(ctx).
		Preload("SaleUnits").
		Find(&salesModels, "sold_at BETWEEN ? AND ?", from, to).
		Error; err != nil {
		return nil, fmt.Errorf("get sales from database: %w", err)
	}

	sales := make([]Sale, 0, len(salesModels))
	for _, s := range salesModels {
		sales = append(sales, FromDBSale(s))
	}

	return sales, nil
}

func (d *Database) GetAggregatedSalesBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]AggregatedSale, error) {
	var sales []AggregatedSale
	if err := d.dbCtx(ctx).
		Raw(
			`
SELECT 
	drugs.name AS drug_name,
	sales.amount AS quantity,
	sales.unit AS unit
FROM
	(
		SELECT drug_code, SUM(amount) as amount, unit
		FROM 
			sale_units JOIN sales ON sale_units.invoice_number = sales.invoice_number
		WHERE 
			sales.sold_at BETWEEN ? AND ?
		GROUP BY drug_code, unit
	) sales
	JOIN drugs ON sales.drug_code = drugs.vmedis_code
ORDER BY drugs.name`,
			from,
			to,
		).
		Find(&sales).
		Error; err != nil {
		return nil, fmt.Errorf("get aggregated sales between %s and %s from DB: %w", from, to, err)
	}

	return sales, nil
}

func (d *Database) GetSalesStatisticsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]Statistics, error) {
	var modelStats []models.SaleStatistics
	if err := d.dbCtx(ctx).
		Where("pulled_at BETWEEN ? AND ?", from, to).
		Order("pulled_at ASC").
		Find(&modelStats).
		Error; err != nil {
		return nil, fmt.Errorf("get sales statistics between %s and %s from DB: %w", from, to, err)
	}

	stats := make([]Statistics, 0, len(modelStats))
	for _, s := range modelStats {
		stats = append(stats, FromDBSaleStatistics(s))
	}

	return stats, nil
}

func (d *Database) UpsertVmedisSales(ctx context.Context, vmedisSales []vmedis.Sale) error {
	if len(vmedisSales) == 0 {
		return nil
	}

	dbSales := slices2.Map(vmedisSales, VmedisSaleToDBSale)

	return d.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "invoice_number"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"updated_at",
					"vmedis_id",
					"sold_at",
					"patient_name",
					"doctor",
					"payment",
					"total",
				}),
			}).
			Omit("SaleUnits").
			Create(&dbSales).
			Error; err != nil {
			return fmt.Errorf("create sales: %w", err)
		}

		for _, sale := range dbSales {
			if len(sale.SaleUnits) > 0 {
				if err := tx.Clauses(
					clause.OnConflict{
						Columns: []clause.Column{{Name: "invoice_number"}, {Name: "id_in_sale"}},
						DoUpdates: clause.AssignmentColumns([]string{
							"updated_at",
							"drug_code",
							"drug_name",
							"batch",
							"amount",
							"unit",
							"unit_price",
							"price_category",
							"discount",
							"tuslah",
							"embalase",
							"total",
						}),
					}).
					Create(&sale.SaleUnits).
					Error; err != nil {
					return fmt.Errorf("create sale units: %w", err)
				}
			} else {
				log.Printf("sale %s has no sale unit", sale.InvoiceNumber)
			}
		}

		return nil
	})
}

func (d *Database) InsertSalesStatistics(ctx context.Context, stats Statistics) error {
	statsModel := stats.ToDBSaleStatistics()
	if err := d.dbCtx(ctx).Create(&statsModel).Error; err != nil {
		return fmt.Errorf("insert sales statistics: %w", err)
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
