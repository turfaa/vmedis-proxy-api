package proxy

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetDailySalesStatistics handles the request to get the daily sales statistics.
func (s *ApiServer) HandleGetDailySalesStatistics(c *gin.Context) {
	var modelStats []models.SaleStatistics
	if err := s.DB.
		Where("pulled_at >= ?", beginningOfToday().Add(30*time.Minute)).
		Order("pulled_at ASC").
		Find(&modelStats).
		Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get historical statistics from DB: " + err.Error(),
		})
		return
	}

	latestStat, err := s.Client.GetDailySalesStatistics(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get latest statistics from API: " + err.Error(),
		})
		return
	}

	var stats []schema.SaleStatistics
	for _, s := range modelStats {
		stats = append(stats, schema.FromModelsSaleStatistics(s))
	}

	latest, err := schema.FromSalesStatisticsClientSchema(time.Now(), latestStat)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to convert latest statistics from API: " + err.Error(),
		})
		return
	}

	if len(stats) == 0 || (stats[len(stats)-1].TotalSales <= latest.TotalSales && stats[len(stats)-1].PulledAt.Before(latest.PulledAt)) {
		stats = append(stats, latest)
	}

	c.JSON(200, schema.SaleStatisticsResponse{History: stats})
}
