package proxy

import (
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"gorm.io/gorm"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	Client            *vmedis.Client
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
				s.HandleGetSales,
			)

			sales.GET(
				"/drugs",
				s.HandleGetSoldDrugs,
			)

			sales.POST(
				"/dump",
				s.HandleDumpSales,
			)

			statistics := sales.Group("/statistics")
			{
				statistics.GET(
					"",
					cache.CacheByRequestURI(store, time.Minute),
					s.HandleGetSalesStatistics,
				)

				statistics.GET(
					"/daily",
					cache.CacheByRequestURI(store, time.Minute),
					s.HandleGetDailySalesStatistics,
				)
			}
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

			procurement.GET(
				"/invoice-calculators",
				cache.CacheByRequestURI(store, time.Hour),
				s.HandleGetInvoiceCalculators,
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
				s.HandleGetDrugsToStockOpname,
			)

			drugs.POST(
				"/dump",
				s.HandleDumpDrugs,
			)
		}

		stockOpnames := v1.Group("/stock-opnames")
		{
			stockOpnames.GET(
				"",
				s.HandleGetStockOpnames,
			)

			stockOpnames.GET(
				"/summaries",
				s.HandleGetStockOpnameSummaries,
			)

			stockOpnames.POST(
				"/dump",
				s.HandleDumpStockOpnames,
			)
		}

		users := v1.Group("/users")
		{
			users.POST(
				"/login",
				s.HandleLogin,
			)
		}
	}
}
