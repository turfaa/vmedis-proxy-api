package proxy

import (
	"time"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	Client      *client.Client
	DB          *gorm.DB
	RedisClient *redis.Client
}

// GinEngine returns the gin engine of the proxy api server.
func (s *ApiServer) GinEngine() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	s.SetupRoute(&r.RouterGroup)
	return r
}

// SetupRoute sets up the routes of the proxy api server.
func (s *ApiServer) SetupRoute(router *gin.RouterGroup) {
	store := CompressedCache{Store: persist.NewRedisStore(s.RedisClient)}

	v1 := router.Group("/api/v1")
	{
		v1.GET(
			"/sales/statistics/daily",
			cache.CacheByRequestURI(store, time.Minute),
			s.HandleGetDailySalesStatistics,
		)

		v1.GET(
			"/procurement/recommendations",
			s.HandleProcurementRecommendations,
		)

		v1.POST(
			"/procurement/recommendations/dump",
			s.HandleDumpProcurementRecommendations,
		)

		v1.GET(
			"/drugs",
			cache.CacheByRequestURI(store, time.Hour),
			s.HandleGetDrugs,
		)
	}
}
