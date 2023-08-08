package proxy

import (
	"time"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	Client *client.Client
	DB     *gorm.DB
}

// GinEngine returns the gin engine of the proxy api server.
func (s *ApiServer) GinEngine() *gin.Engine {
	r := gin.Default()
	s.SetupRoute(&r.RouterGroup)
	return r
}

// SetupRoute sets up the routes of the proxy api server.
func (s *ApiServer) SetupRoute(router *gin.RouterGroup) {
	store := persist.NewMemoryStore(time.Minute)

	v1 := router.Group("/api/v1")
	{
		v1.GET(
			"/sales/statistics/daily",
			cache.CacheByRequestURI(store, time.Minute),
			s.HandleGetDailySalesStatistics,
		)
	}
}

// HandleGetDailySalesStatistics handles the request to get the daily sales statistics.
func (s *ApiServer) HandleGetDailySalesStatistics(c *gin.Context) {
	var modelStats []models.SaleStatistics
	if err := s.DB.
		Where("pulled_at >= ?", time.Now().UTC().Truncate(time.Hour*24).Add(time.Hour)).
		Order("pulled_at ASC").
		Find(&modelStats).
		Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get historical statistics from DB: " + err.Error(),
		})
		return
	}

	latestStat, err := s.Client.GetDailySalesStatistics()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get latest statistics from API: " + err.Error(),
		})
		return
	}

	var stats []SaleStatistics
	for _, s := range modelStats {
		stats = append(stats, FromModelsSaleStatistics(s))
	}

	latest, err := FromSalesStatisticsClientSchema(time.Now(), latestStat)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to convert latest statistics from API: " + err.Error(),
		})
		return
	}

	if len(stats) == 0 || (stats[len(stats)-1].TotalSales <= latest.TotalSales && stats[len(stats)-1].PulledAt.Before(latest.PulledAt)) {
		stats = append(stats, latest)
	}

	c.JSON(200, SaleStatisticsResponse{History: stats})
}
