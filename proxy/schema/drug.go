package schema

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
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

// FromClientDrug converts Drug from client schema to proxy schema.
func FromClientDrug(cd vmedis.Drug) Drug {
	units := make([]Unit, 0, len(cd.Units))
	for _, cu := range cd.Units {
		units = append(units, FromClientUnit(cu))
	}

	stocks := make([]Stock, 0, len(cd.Stocks))
	for _, cs := range cd.Stocks {
		stocks = append(stocks, FromClientStock(cs))
	}

	return Drug{
		VmedisCode:   cd.VmedisCode,
		Name:         cd.Name,
		Manufacturer: cd.Manufacturer,
		Supplier:     cd.Supplier,
		MinimumStock: FromClientStock(cd.MinimumStock),
		Units:        units,
		Stocks:       stocks,
	}
}

// FromClientUnit converts Unit from client schema to proxy schema.
func FromClientUnit(cu vmedis.Unit) Unit {
	return Unit{
		Unit:                   cu.Unit,
		ParentUnit:             cu.ParentUnit,
		ConversionToParentUnit: cu.ConversionToParentUnit,
		UnitOrder:              cu.UnitOrder,
		PriceOne:               cu.PriceOne,
		PriceTwo:               cu.PriceTwo,
		PriceThree:             cu.PriceThree,
	}
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
