package procurement

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type ApiHandler struct {
	service *Service
}

func (h *ApiHandler) GetRecommendations(c *gin.Context) {
	recommendations, err := h.service.GetRecommendations(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get recommendations: %s", err),
		})
		return
	}

	c.JSON(200, recommendations)
}

func (h *ApiHandler) DumpRecommendationsFromVmedisToRedis(c *gin.Context) {
	go func() {
		if err := h.service.DumpRecommendationsFromVmedisToRedis(context.Background()); err != nil {
			log.Printf("Failed to dump recommendations from vmedis to redis: %s", err)
		}
	}()
	
	c.JSON(200, gin.H{
		"message": "dumping recommendations from vmedis to redis",
	})
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
