package schema

import (
	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// DrugsResponse is the response of the drugs endpoint.
type DrugsResponse struct {
	Drugs []Drug `json:"drugs"`
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisCode   string `json:"vmedisCode"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Supplier     string `json:"supplier"`
	MinimumStock Stock  `json:"minimumStock"`
}

// FromClientDrug converts Drug from client schema to proxy schema.
func FromClientDrug(cd client.Drug) Drug {
	return Drug{
		VmedisCode:   cd.VmedisCode,
		Name:         cd.Name,
		Manufacturer: cd.Manufacturer,
		Supplier:     cd.Supplier,
		MinimumStock: FromClientStock(cd.MinimumStock),
	}
}

// FromModelsDrug converts Drug from models.Drug to proxy schema.
func FromModelsDrug(drug models.Drug) Drug {
	return Drug{
		VmedisCode:   drug.VmedisCode,
		Name:         drug.Name,
		Manufacturer: drug.Manufacturer,
		MinimumStock: FromModelsStock(drug.MinimumStock),
	}
}
