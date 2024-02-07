package models

import (
	"time"

	"gorm.io/datatypes"
)

// StockOpname represents a stock opname.
type StockOpname struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time

	VmedisID            string         `gorm:"unique"`
	Date                datatypes.Date `gorm:"index:idx_so_date_drug_code"`
	DrugCode            string         `gorm:"index:idx_so_date_drug_code;index:idx_so_drug_code"`
	DrugName            string
	BatchCode           string
	Unit                string
	InitialQuantity     float64
	RealQuantity        float64
	QuantityDifference  float64
	HPPDifference       float64
	SalePriceDifference float64
	Notes               string
}
