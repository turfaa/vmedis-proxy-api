package schema

import (
	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// DrugsResponse is the response of the drugs endpoint.
type DrugsResponse struct {
	Drugs []Drug `json:"drugs"`
	Date  string `json:"date,omitempty"`
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisCode   string `json:"vmedisCode,omitempty"`
	Name         string `json:"name,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Supplier     string `json:"supplier,omitempty"`
	MinimumStock Stock  `json:"minimumStock"`
	Units        []Unit `json:"units,omitempty"`
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
func FromClientDrug(cd client.Drug) Drug {
	var units []Unit
	for _, cu := range cd.Units {
		units = append(units, FromClientUnit(cu))
	}

	return Drug{
		VmedisCode:   cd.VmedisCode,
		Name:         cd.Name,
		Manufacturer: cd.Manufacturer,
		Supplier:     cd.Supplier,
		MinimumStock: FromClientStock(cd.MinimumStock),
		Units:        units,
	}
}

// FromClientUnit converts Unit from client schema to proxy schema.
func FromClientUnit(cu client.Unit) Unit {
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
	var units []Unit
	for _, mu := range drug.Units {
		units = append(units, FromModelsUnit(mu))
	}

	return Drug{
		VmedisCode:   drug.VmedisCode,
		Name:         drug.Name,
		Manufacturer: drug.Manufacturer,
		MinimumStock: FromModelsStock(drug.MinimumStock),
		Units:        units,
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
