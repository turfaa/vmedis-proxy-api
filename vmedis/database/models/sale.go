package models

import "time"

// Sale represents a sale.
type Sale struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// VmedisID is the ID of the sale in Vmedis.
	VmedisID      int       `gorm:"index"`
	SoldAt        time.Time `gorm:"index"`
	InvoiceNumber string    `gorm:"unique"`
	PatientName   string    `gorm:"index"`
	Doctor        string    `gorm:"index"`
	Payment       string    `gorm:"index"`
	Total         float64
	SaleUnits     []SaleUnit `gorm:"foreignKey:InvoiceNumber;references:InvoiceNumber"`
}

// SaleUnit represents one unit of a drug in a sale.
type SaleUnit struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	InvoiceNumber string `gorm:"uniqueIndex:idx_sale_unit_invoice_number_id_in_sale"`
	IDInSale      int    `gorm:"uniqueIndex:idx_sale_unit_invoice_number_id_in_sale"`
	DrugCode      string `gorm:"index"`
	DrugName      string `gorm:"index"`
	Batch         string
	Amount        float64
	Unit          string
	UnitPrice     float64
	PriceCategory string
	Discount      float64
	Tuslah        float64
	Embalase      float64
	Total         float64
}
