package proxy

import (
	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/proxy/schema"
)

// HandleGetInvoiceCalculators handles GET /procurement/invoice-calculators.
func (s *ApiServer) HandleGetInvoiceCalculators(c *gin.Context) {
	var calculators []models.InvoiceCalculator
	if err := s.db.Preload("Components").Order("supplier").Find(&calculators).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var res schema.InvoiceCalculatorsResponse
	for _, calculator := range calculators {
		res.Calculators = append(res.Calculators, schema.FromModelsInvoiceCalculator(calculator))
	}

	c.JSON(200, res)
}