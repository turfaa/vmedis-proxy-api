package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Procurement struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// InvoiceNumber stays unique across soft-deleted procurements too, so
	// re-dumping a procurement that was soft-deleted here updates the
	// soft-deleted row instead of resurrecting it as a second, visible one.
	InvoiceNumber          string         `gorm:"unique"`
	InvoiceDate            datatypes.Date `gorm:"index"`
	InputDate              time.Time
	Supplier               string `gorm:"index"`
	Warehouse              string
	PaymentType            string `gorm:"index"`
	PaymentAccount         string
	Operator               string
	CashDiscountPercentage float64
	DiscountPercentage     float64
	DiscountAmount         float64
	TaxPercentage          float64
	TaxAmount              float64
	MiscellaneousCost      float64
	Total                  float64
	ProcurementUnits       []ProcurementUnit `gorm:"foreignKey:InvoiceNumber;references:InvoiceNumber"`
}

type ProcurementUnit struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	InvoiceNumber           string `gorm:"index;uniqueIndex:idx_procurement_unit_invoice_number_id_in_procurement"`
	IDInProcurement         int    `gorm:"uniqueIndex:idx_procurement_unit_invoice_number_id_in_procurement"`
	DrugCode                string `gorm:"index"`
	DrugName                string `gorm:"index"`
	Amount                  float64
	Unit                    string
	UnitBasePrice           float64
	DiscountPercentage      float64
	DiscountTwoPercentage   float64
	DiscountThreePercentage float64
	TotalUnitPrice          float64
	UnitTaxedPrice          float64
	ExpiryDate              time.Time
	BatchNumber             string
	Total                   float64
}
