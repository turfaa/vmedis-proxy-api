package models

import (
	"time"

	"gorm.io/gorm"
)

// Supplier is a supplier the pharmacy procures drugs from, dumped from vmedis.
type Supplier struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// VmedisID is the ID of the supplier in vmedis. It stays unique across
	// soft-deleted suppliers too, so re-dumping a supplier that was
	// soft-deleted here revives it.
	VmedisID          int    `gorm:"unique"`
	Code              string `gorm:"index"`
	Name              string `gorm:"index"`
	Address           string
	City              string
	Phone             string
	MobilePhone       string
	BankAccountNumber string
	// TaxNumber is the supplier's NPWP.
	TaxNumber    string
	Active       bool
	LeadTimeDays int
	Rating       float64
}
