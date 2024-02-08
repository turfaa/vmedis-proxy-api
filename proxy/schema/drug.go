package schema

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// Drug is a drug in the inventory.
type Drug struct {
	VmedisCode   string  `json:"vmedisCode,omitempty"`
	Name         string  `json:"name,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	Supplier     string  `json:"supplier,omitempty"`
	MinimumStock Stock   `json:"minimumStock"`
	Units        []Unit  `json:"units"`
	Stocks       []Stock `json:"stocks"`
}

// Unit is a unit of a drug.
type Unit struct {
	Unit                   string  `json:"unit,omitempty"`
	ParentUnit             string  `json:"parentUnit,omitempty"`
	ConversionToParentUnit float64 `json:"conversionToParentUnit,omitempty"`

	// UnitOrder is the order of the unit of the drug.
	// The smallest unit has the lowest order.
	UnitOrder int `json:"unitOrder"`

	// PriceOne, PriceTwo, and PriceThree are the prices of the drug for different segments.
	// PriceOne is the price for common customers.
	// PriceTwo is the price for medical facilities.
	// PriceThree is the price for prescription.
	PriceOne   float64 `json:"priceOne,omitempty"`
	PriceTwo   float64 `json:"priceTwo,omitempty"`
	PriceThree float64 `json:"priceThree,omitempty"`
}

// FromModelsDrug converts Drug from models.Drug to proxy schema.
func FromModelsDrug(drug models.Drug) Drug {
	units := make([]Unit, 0, len(drug.Units))
	for _, mu := range drug.Units {
		units = append(units, FromModelsUnit(mu))
	}

	stocks := make([]Stock, 0, len(drug.Stocks))
	for _, ms := range drug.Stocks {
		stocks = append(stocks, FromModelsDrugStock(ms))
	}

	return Drug{
		VmedisCode:   drug.VmedisCode,
		Name:         drug.Name,
		Manufacturer: drug.Manufacturer,
		MinimumStock: FromModelsStock(drug.MinimumStock),
		Units:        units,
		Stocks:       stocks,
	}
}

// FromModelsUnit converts Unit from models.DrugUnit to proxy schema.
func FromModelsUnit(unit models.DrugUnit) Unit {
	return Unit{
		Unit:                   unit.Unit,
		ParentUnit:             unit.ParentUnit,
		ConversionToParentUnit: unit.ConversionToParentUnit,
		UnitOrder:              unit.UnitOrder,
		PriceOne:               unit.PriceOne,
		PriceTwo:               unit.PriceTwo,
		PriceThree:             unit.PriceThree,
	}
}
