package proxy

import (
	"time"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	Client            *client.Client
	DB                *gorm.DB
	RedisClient       *redis.Client
	DrugDetailsPuller chan<- models.Drug
}

// GinEngine returns the gin engine of the proxy api server.
func (s *ApiServer) GinEngine() *gin.Engine {
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default())

	s.SetupRoute(&r.RouterGroup)
	return r
}

// SetupRoute sets up the routes of the proxy api server.
func (s *ApiServer) SetupRoute(router *gin.RouterGroup) {
	store := CompressedCache{Store: persist.NewRedisStore(s.RedisClient)}

	v1 := router.Group("/api/v1")
	{
		sales := v1.Group("/sales")
		{
			sales.GET(
				"",
				cache.CacheByRequestURI(store, 10*time.Minute),
				s.HandleGetSales,
			)

			sales.GET(
				"/statistics/daily",
				cache.CacheByRequestURI(store, time.Minute),
				s.HandleGetDailySalesStatistics,
			)

			sales.GET(
				"/drugs",
				cache.CacheByRequestURI(store, 10*time.Minute),
				s.HandleGetSoldDrugs,
			)

			sales.POST(
				"/dump",
				s.HandleDumpSales,
			)
		}

		procurement := v1.Group("/procurement")
		{
			procurement.GET(
				"/recommendations",
				s.HandleProcurementRecommendations,
			)

			procurement.POST(
				"/recommendations/dump",
				s.HandleDumpProcurementRecommendations,
			)
		}

		drugs := v1.Group("/drugs")
		{
			drugs.GET(
				"",
				cache.CacheByRequestURI(store, time.Hour),
				s.HandleGetDrugs,
			)

			drugs.GET(
				"/to-stock-opname",
				cache.CacheByRequestURI(store, time.Hour),
				s.HandleGetDrugsToStockOpname,
			)

			drugs.POST(
				"/dump",
				s.HandleDumpDrugs,
			)
		}

		stockOpnames := v1.Group("/stock-opnames")
		{
			stockOpnames.POST(
				"/dump",
				s.HandleDumpStockOpnames,
			)
		}
	}
}
