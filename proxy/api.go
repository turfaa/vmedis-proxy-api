package proxy

import (
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	gzip "github.com/turfaa/gin-gzip"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/auth"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/sale"
	"github.com/turfaa/vmedis-proxy-api/shift"
	"github.com/turfaa/vmedis-proxy-api/stockopname"
)

// ApiServer is the proxy api server.
type ApiServer struct {
	db          *gorm.DB
	redisClient *redis.Client
	authService *auth.Service

	authHandler        *auth.ApiHandler
	drugHandler        *drug.ApiHandler
	saleHandler        *sale.ApiHandler
	procurementHandler *procurement.ApiHandler
	stockOpnameHandler *stockopname.ApiHandler
	shiftHandler       *shift.ApiHandler
}

// GinEngine returns the gin engine of the proxy api server.
func (s *ApiServer) GinEngine() *gin.Engine {
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default())
	r.Use(auth.GinMiddleware(s.authService))

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
				s.saleHandler.GetSales,
			)

			sales.GET(
				"/drugs",
				s.saleHandler.GetSoldDrugs,
			)

			sales.GET(
				"/statistics/sensors",
				cache.CacheByRequestURI(store, time.Minute),
				s.saleHandler.GetSalesStatisticsSensors,
			)

			sales.GET(
				"/statistics",
				cache.CacheByRequestURI(store, time.Minute),
				s.saleHandler.GetSalesStatistics,
			)

			sales.POST(
				"/dump",
				s.saleHandler.DumpTodaySales,
			)
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
				s.authHandler.Login,
			)
		}

		authGroup := v1.Group("/auth")
		{
			authGroup.POST(
				"/login",
				s.authHandler.Login,
			)
		}
	}

	// v2 is the new api version where:
	// 	- Response is already generated on the server side.
	// 	- Authentication might affect the response.
	v2 := router.Group("/api/v2")
	{
		drugs := v2.Group("/drugs")
		{
			drugs.GET(
				"",
				cache.Cache(store, time.Minute, cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, cache.Strategy) {
					return true, cache.Strategy{
						CacheKey: c.Request.RequestURI + "$$" + string(auth.FromGinContext(c).Role),
					}
				})),
				s.drugHandler.GetDrugsV2,
			)
		}

		procurements := v2.Group("/procurements")
		{
			procurements.GET(
				"/drugs/:drug_code/last",
				auth.AllowedRoles(auth.RoleAdmin, auth.RoleStaff),
				cache.CacheByRequestURI(store, time.Minute),
				s.procurementHandler.GetLastDrugProcurements,
			)
		}

		shifts := v2.Group("/shifts")
		{
			shifts.GET(
				"",
				auth.AllowedRoles(auth.RoleAdmin, auth.RoleStaff),
				s.shiftHandler.GetShifts,
			)

			shifts.POST(
				"/dump",
				auth.AllowedRoles(auth.RoleAdmin, auth.RoleStaff),
				s.shiftHandler.DumpShiftsFromVmedisToDB,
			)

			shifts.GET(
				"/:vmedis_id",
				auth.AllowedRoles(auth.RoleAdmin, auth.RoleStaff),
				s.shiftHandler.GetShiftByVmedisID,
			)

			shifts.GET(
				"/:vmedis_id/show",
				auth.AllowedRoles(auth.RoleAdmin, auth.RoleStaff),
				s.shiftHandler.ShowShift,
			)
		}
	}
}

// NewApiServer creates a new api server.
func NewApiServer(
	db *gorm.DB,
	redisClient *redis.Client,
	authService *auth.Service,

	authHandler *auth.ApiHandler,
	drugHandler *drug.ApiHandler,
	saleHandler *sale.ApiHandler,
	procurementHandler *procurement.ApiHandler,
	stockOpnameHandler *stockopname.ApiHandler,
	shiftHandler *shift.ApiHandler,
) *ApiServer {
	return &ApiServer{
		db:          db,
		redisClient: redisClient,
		authService: authService,

		authHandler:        authHandler,
		drugHandler:        drugHandler,
		saleHandler:        saleHandler,
		procurementHandler: procurementHandler,
		stockOpnameHandler: stockOpnameHandler,
		shiftHandler:       shiftHandler,
	}
}
