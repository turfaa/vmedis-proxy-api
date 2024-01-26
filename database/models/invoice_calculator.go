package models

import "time"

// InvoiceCalculator represents an invoice calculator for a supplier.
type InvoiceCalculator struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Supplier    string `gorm:"unique"`
	ShouldRound bool
	Components  []InvoiceComponent `gorm:"foreignKey:Supplier;references:Supplier"`
}

// InvoiceComponent represents an invoice component.
type InvoiceComponent struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Supplier   string `gorm:"uniqueIndex:idx_invoice_component_supplier_name"`
	Name       string `gorm:"uniqueIndex:idx_invoice_component_supplier_name"`
	Multiplier float64
}
