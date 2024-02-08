package procurement

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/gin2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type ApiHandler struct {
	service *Service
}

func (h *ApiHandler) DumpProcurements(c *gin.Context) {
	var request DumpProcurementsRequest
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid request: %s", err),
		})
		return
	}

	startDate, endDate, err := request.ExtractDates()
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid request, failed to extract dates: %s", err),
		})
		return
	}

	go func() {
		if err := h.service.DumpProcurementsBetweenDatesFromVmedisToDB(context.Background(), startDate, endDate); err != nil {
			log.Printf("Failed to dump procurements from vmedis to DB: %s", err)
		}
	}()

	c.JSON(200, gin.H{
		"message": "dumping procurements from vmedis to DB",
	})
}

func (h *ApiHandler) GetRecommendations(c *gin.Context) {
	recommendations, err := h.service.GetRecommendations(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get recommendations: %s", err),
		})
		return
	}

	// We need to use the experimental JSON renderer to make use of the new "inline" behaviour.
	c.Render(200, gin2.ExperimentalJSONRenderer{Data: recommendations})
}

func (h *ApiHandler) DumpRecommendations(c *gin.Context) {
	go func() {
		if err := h.service.DumpRecommendationsFromVmedisToRedis(context.Background()); err != nil {
			log.Printf("Failed to dump recommendations from vmedis to redis: %s", err)
		}
	}()

	c.JSON(200, gin.H{
		"message": "dumping recommendations from vmedis to redis",
	})
}

func (h *ApiHandler) GetInvoiceCalculators(c *gin.Context) {
	calculators, err := h.service.GetInvoiceCalculators(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get invoice calculators: %s", err),
		})
		return
	}

	c.JSON(200, InvoiceCalculatorsResponse{Calculators: calculators})
}

func NewApiHandler(
	db *gorm.DB,
	redisClient *redis.Client,
	vmedisClient *vmedis.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) *ApiHandler {
	return &ApiHandler{
		service: &Service{
			db:              NewDatabase(db),
			redisDB:         NewRedisDatabase(redisClient),
			vmedis:          vmedisClient,
			drugProducer:    drugProducer,
			drugUnitsGetter: drugUnitsGetter,
		},
	}
}
