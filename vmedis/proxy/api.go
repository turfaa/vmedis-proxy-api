package proxy

import (
	"time"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	Client *client.Client
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
	stats, err := s.Client.GetDailySalesStatistics()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, stats)
}
