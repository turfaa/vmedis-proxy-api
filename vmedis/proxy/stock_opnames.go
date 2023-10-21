package proxy

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetStockOpnames handles the request to get the stock opnames.
func (s *ApiServer) HandleGetStockOpnames(c *gin.Context) {
	dayFrom, _, err := getTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "failed to parse date query: " + err.Error(),
		})
		return
	}

	var stockOpnames []models.StockOpname
	if err := s.DB.Where("date = ?", datatypes.Date(dayFrom)).Order("vmedis_id").Find(&stockOpnames).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get stock opnames: " + err.Error(),
		})
		return
	}

	sos := make([]schema.StockOpname, len(stockOpnames))
	for i, so := range stockOpnames {
		sos[i] = schema.FromModelsStockOpname(so)
	}

	c.JSON(200, schema.StockOpnamesResponse{StockOpnames: sos, Date: dayFrom.Format("2006-01-02")})
}

// HandleDumpStockOpnames handles the request to dump the stock opnames.
func (s *ApiServer) HandleDumpStockOpnames(c *gin.Context) {
	go dumper.DumpDailyStockOpnames(context.Background(), s.DB, s.Client)
	c.JSON(200, gin.H{
		"message": "dumping stock opnames",
	})
}
