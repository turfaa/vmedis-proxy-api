package models

import (
	"time"
)

// Drug represents a drug.
type Drug struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	VmedisID     int    `gorm:"unique"`
	VmedisCode   string `gorm:"unique"`
	Name         string
	Manufacturer string `gorm:"index"`
	MinimumStock Stock  `gorm:"embedded;embeddedPrefix:minimum_stock_"`

	Units []DrugUnit `gorm:"foreignKey:DrugVmedisCode;references:VmedisCode"`
}

// SmallestUnit returns the smallest unit of the drug.
// The smallest unit is the unit that has no parent unit.
// If the drug has no units, it returns false.
// If the drug has more than one smallest unit, it returns any one of them.
func (d Drug) SmallestUnit() (smallestUnit DrugUnit, found bool) {
	if len(d.Units) == 0 {
		return
	}

	for _, unit := range d.Units {
		if unit.ParentUnit == "" {
			smallestUnit = unit
			found = true
			return
		}
	}

	return
}

// DrugUnit represents a drug unit and its relation to another unit of the drug.
type DrugUnit struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	DrugVmedisCode         string `gorm:"uniqueIndex:idx_drug_units_code_unit"`
	Unit                   string `gorm:"index;uniqueIndex:idx_drug_units_code_unit"`
	ParentUnit             string
	ConversionToParentUnit float64

	// UnitOrder is the order of the unit of the drug.
	// The smallest unit has the lowest order.
	UnitOrder int

	// PriceOne, PriceTwo, and PriceThree are the prices of the drug for different segments.
	// PriceOne is the price for common customers.
	// PriceTwo is the price for medical facilities.
	// PriceThree is the price for prescription.
	PriceOne   float64
	PriceTwo   float64
	PriceThree float64
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string
	Quantity float64
}
