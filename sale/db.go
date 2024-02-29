package sale

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
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

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}
