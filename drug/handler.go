package drug

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/jsonpb"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	"github.com/turfaa/vmedis-proxy-api/time2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// ApiHandler is the handler for drug-related APIs.
type ApiHandler struct {
	service *Service
}

// GetDrugs handles requests to get all drugs.
func (h *ApiHandler) GetDrugs(c *gin.Context) {
	drugs, err := h.service.GetDrugs(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs: %s", err),
		})
		return
	}

	for i, drug := range drugs {
		drug.Units = filterUnits(drug.Units)
		drugs[i] = drug
	}

	c.JSON(200, DrugsResponse{Drugs: drugs})
}

// GetDrugsToStockOpname handles requests to get the drugs to stock opname.
func (h *ApiHandler) GetDrugsToStockOpname(c *gin.Context) {
	mode := strings.ToLower(c.DefaultQuery("mode", "sales-based"))

	switch mode {
	case "conservative":
		h.GetConservativeDrugsToStockOpname(c)

	case "sales-based":
		h.GetSalesBasedDrugsToStockOpname(c)

	default:
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid mode: %s", mode),
		})
	}
}

// GetConservativeDrugsToStockOpname handles requests to get the drugs to stock opname based on all drugs.
func (h *ApiHandler) GetConservativeDrugsToStockOpname(c *gin.Context) {
	_, todayUntil, err := time2.GetOneDayFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse date query: %s", err),
		})
		return
	}

	startFrom, yesterdayUntil := time.Date(2024, 01, 06, 0, 0, 0, 0, time.Local), todayUntil.Add(-24*time.Hour)

	drugs, err := h.service.GetDrugsToStockOpname(c, startFrom, yesterdayUntil)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs to stock opname: %s", err),
		})
		return
	}

	c.JSON(200, DrugsResponse{Drugs: drugs})
}

// GetSalesBasedDrugsToStockOpname handles requests to get the drugs to stock opname based on sales in the last month.
func (h *ApiHandler) GetSalesBasedDrugsToStockOpname(c *gin.Context) {
	_, todayUntil, err := time2.GetOneDayFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse date query: %s", err),
		})
		return
	}

	startFrom, yesterdayUntil := time.Date(2024, 01, 06, 0, 0, 0, 0, time.Local), todayUntil.Add(-24*time.Hour)

	drugs, err := h.service.GetSalesBasedDrugsToStockOpname(c, startFrom, yesterdayUntil)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get sales-based drugs to stock opname: %s", err),
		})
		return
	}

	c.JSON(200, DrugsResponse{Drugs: drugs})
}

// DumpDrugs handles requests to dump the drugs.
func (h *ApiHandler) DumpDrugs(c *gin.Context) {
	go func() {
		if err := h.service.DumpDrugsFromVmedisToDB(context.Background()); err != nil {
			log.Printf("failed to dump drugs: %s", err)
		}
	}()

	c.JSON(200, gin.H{
		"message": "dumping drugs",
	})
}

// NewApiHandler creates a new ApiHandler.
func NewApiHandler(db *gorm.DB, vmedisClient *vmedis.Client, kafkaWriter *kafka.Writer) *ApiHandler {
	return &ApiHandler{
		service: NewService(db, vmedisClient, kafkaWriter),
	}
}

type ConsumerHandler struct {
	service *Service
	cache   *Cache
}

func (h *ConsumerHandler) DumpDrugDetailsByVmedisCode(ctx context.Context, kafkaMessage kafka.Message) error {
	var payload kafkapb.UpdatedDrugByVmedisCode
	if err := jsonpb.UnmarshalString(string(kafkaMessage.Value), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message: %s", err)
	}

	processed, err := h.cache.HasDrugDetailsByVmedisCodeProcessed(ctx, payload.RequestKey)
	if err != nil {
		return fmt.Errorf("failed to check if drug details by vmedis code processed: %s", err)
	}

	if processed {
		return nil
	}

	log.Printf("Processing message %s to dump drug details by vmedis code", payload.RequestKey)

	if err := h.service.DumpDrugDetailsFromVmedisToDBByVmedisCode(ctx, payload.VmedisCode); err != nil {
		return fmt.Errorf("failed to dump drug details by vmedis code: %s", err)
	}

	if err := h.cache.MarkDrugDetailsByVmedisCodeProcessed(ctx, payload.RequestKey); err != nil {
		return fmt.Errorf("failed to mark drug details by vmedis code processed: %s", err)
	}

	return nil
}

func (h *ConsumerHandler) DumpDrugDetailsByVmedisID(ctx context.Context, kafkaMessage kafka.Message) error {
	var payload kafkapb.UpdatedDrugByVmedisID
	if err := jsonpb.UnmarshalString(string(kafkaMessage.Value), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message: %s", err)
	}

	processed, err := h.cache.HasDrugDetailsByVmedisIDProcessed(ctx, payload.RequestKey)
	if err != nil {
		return fmt.Errorf("failed to check if drug details by vmedis id processed: %s", err)
	}

	if processed {
		return nil
	}

	log.Printf("Processing message %s to dump drug details by vmedis id", payload.RequestKey)

	if err := h.service.DumpDrugDetailsFromVmedisToDBByVmedisID(ctx, payload.VmedisId); err != nil {
		return fmt.Errorf("failed to dump drug details by vmedis id: %s", err)
	}

	if err := h.cache.MarkDrugDetailsByVmedisIDProcessed(ctx, payload.RequestKey); err != nil {
		return fmt.Errorf("failed to mark drug details by vmedis id processed: %s", err)
	}

	return nil
}

// NewConsumerHandler creates a new ConsumerHandler.
func NewConsumerHandler(db *gorm.DB, redisClient *redis.Client, vmedisClient *vmedis.Client, kafkaWriter *kafka.Writer) *ConsumerHandler {
	return &ConsumerHandler{
		service: NewService(db, vmedisClient, kafkaWriter),
		cache:   NewCache(redisClient),
	}
}
