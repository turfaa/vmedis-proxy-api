package proxy

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetSales handles the GET /sales endpoint.
func (s *ApiServer) HandleGetSales(c *gin.Context) {
	date := c.Query("date")

	var from, until time.Time
	if date == "" {
		from, until = today()
	} else {
		var err error
		from, until, err = day(date)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "failed to parse date: " + err.Error(),
			})
			return
		}
	}

	var salesModels []models.Sale
	if err := s.DB.Preload("SaleUnits").Find(&salesModels, "sold_at >= ? AND sold_at <= ?", from, until).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get sales from database: " + err.Error(),
		})
		return
	}

	sales := make([]schema.Sale, len(salesModels))
	for i, sale := range salesModels {
		sales[i] = schema.FromModelsSale(sale)
	}

	c.JSON(200, schema.SalesResponse{Sales: sales})
}

// HandleDumpSales handles the request to dump today's sales.
func (s *ApiServer) HandleDumpSales(c *gin.Context) {
	go dumper.DumpDailySales(context.Background(), s.DB, s.Client)

	c.JSON(200, gin.H{
		"message": "dumping today's sales",
	})
}
