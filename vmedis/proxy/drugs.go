package proxy

import (
	"github.com/gin-gonic/gin"

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

// HandleGetDrugs handles the request to get the drugs.
func (s *ApiServer) HandleGetDrugs(c *gin.Context) {
	var drugs []models.Drug
	if err := s.DB.Find(&drugs).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get drugs from database: " + err.Error(),
		})
		return
	}

	var res DrugsResponse
	for _, drug := range drugs {
		res.Drugs = append(res.Drugs, FromModelsDrug(drug))
	}

	c.JSON(200, res)
}
