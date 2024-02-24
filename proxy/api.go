package proxy

import (
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/stockopname"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	client      *vmedis.Client
	db          *gorm.DB
	redisClient *redis.Client

	drugProducer *drug.Producer

	drugHandler        *drug.ApiHandler
	procurementHandler *procurement.ApiHandler
	stockOpnameHandler *stockopname.ApiHandler
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
	store := CompressedCache{Store: persist.NewRedisStore(s.redisClient)}

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

		// We are migrating to /procurements
		procurementHandlers := v1.Group("/procurement")
		procurementsHandlers := v1.Group("/procurements")

		for _, group := range []*gin.RouterGroup{procurementHandlers, procurementsHandlers} {
			group.GET(
				"/recommendations",
				s.procurementHandler.GetRecommendations,
			)

			group.POST(
				"/recommendations/dump",
				s.procurementHandler.DumpRecommendations,
			)

			group.GET(
				"/invoice-calculators",
				cache.CacheByRequestURI(store, time.Hour),
				s.procurementHandler.GetInvoiceCalculators,
			)

			group.POST(
				"/dump",
				s.procurementHandler.DumpProcurements,
			)
		}

		drugs := v1.Group("/drugs")
		{
			drugs.GET(
				"",
				cache.CacheByRequestURI(store, time.Minute),
				s.drugHandler.GetDrugs,
			)

			drugs.GET(
				"/to-stock-opname",
				s.drugHandler.GetDrugsToStockOpname,
			)

			drugs.POST(
				"/dump",
				s.drugHandler.DumpDrugs,
			)
		}

		stockOpnames := v1.Group("/stock-opnames")
		{
			stockOpnames.GET(
				"",
				s.stockOpnameHandler.GetStockOpnames,
			)

			stockOpnames.GET(
				"/compacted",
				s.stockOpnameHandler.GetCompactedStockOpnames,
			)

			stockOpnames.GET(
				"/summaries",
				s.stockOpnameHandler.GetStockOpnameSummaries,
			)

			stockOpnames.POST(
				"/dump",
				s.stockOpnameHandler.DumpTodayStockOpnames,
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

// NewApiServer creates a new api server.
func NewApiServer(
	client *vmedis.Client,
	db *gorm.DB,
	redisClient *redis.Client,
	kafkaWriter *kafka.Writer,
) *ApiServer {
	drugProducer := drug.NewProducer(kafkaWriter)
	drugDB := drug.NewDatabase(db)

	return &ApiServer{
		client:       client,
		db:           db,
		redisClient:  redisClient,
		drugProducer: drugProducer,

		drugHandler:        drug.NewApiHandler(db, client, kafkaWriter),
		procurementHandler: procurement.NewApiHandler(db, redisClient, client, drugProducer, drugDB),
		stockOpnameHandler: stockopname.NewApiHandler(db, client, drugProducer),
	}
}
