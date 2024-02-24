package stockopname

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/time2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type ApiHandler struct {
	service *Service
}

func (h *ApiHandler) GetStockOpnames(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid time range: %s", err),
		})
		return
	}

	stockOpnames, err := h.service.GetStockOpnamesBetweenTime(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get stock opnames: %s", err),
		})
		return
	}

	c.JSON(200, StockOpnamesResponse{StockOpnames: stockOpnames})
}

func (h *ApiHandler) GetCompactedStockOpnames(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid time range: %s", err),
		})
		return
	}

	stockOpnames, err := h.service.GetCompactedStockOpnamesBetweenTime(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get stock opnames: %s", err),
		})
		return
	}

	c.JSON(200, CompactedStockOpnamesResponse{StockOpnames: stockOpnames})
}

func (h *ApiHandler) GetStockOpnameSummaries(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid time range: %s", err),
		})
		return
	}

	summaries, err := h.service.GetStockOpnameSummariesBetweenTime(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get stock opname summaries: %s", err),
		})
		return
	}

	c.JSON(200, SummariesResponse{Summaries: summaries})
}

func (h *ApiHandler) DumpTodayStockOpnames(c *gin.Context) {
	go func() {
		if err := h.service.DumpTodayStockOpnamesFromVmedisToDB(context.Background()); err != nil {
			log.Printf("Failed to dump stock opnames from Vmedis to DB: %s", err)
		}
	}()

	c.JSON(200, gin.H{
		"message": "dumping stock opnames from vmedis to DB",
	})
}

func NewApiHandler(db *gorm.DB, vmedisClient *vmedis.Client, drugProducer UpdatedDrugProducer) *ApiHandler {
	return &ApiHandler{
		service: NewService(db, vmedisClient, drugProducer),
	}
}
