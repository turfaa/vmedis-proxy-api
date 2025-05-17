package sale

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"
)

type ApiHandler struct {
	service *Service
}

func (s *ApiHandler) GetSales(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	sales, err := s.service.GetSalesBetweenTime(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, SalesResponse{Sales: sales})
}

func (s *ApiHandler) GetSoldDrugs(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	soldDrugs, err := s.service.GetSoldDrugsBetweenTime(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, SoldDrugsResponse{Drugs: soldDrugs})
}

func (s *ApiHandler) GetSalesStatistics(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	stats, err := s.service.GetSalesStatisticsBetweenTime(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(
		200,
		StatisticsResponse{
			History:      stats,
			DailyHistory: GenerateDailyHistory(stats),
		},
	)
}

func (s *ApiHandler) GetSalesStatisticsSensors(c *gin.Context) {
	sensors, err := s.service.GetSalesStatisticsSensors(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(200, sensors.ToStatisticsSensorsResponse())
}

func (s *ApiHandler) DumpTodaySales(c *gin.Context) {
	go func() {
		if err := s.service.DumpTodaySalesStatisticsFromVmedisToDB(context.Background()); err != nil {
			log.Printf("Error dumping today's sales: %s", err)
		}
	}()

	c.JSON(200, gin.H{
		"message": "dumping today's sales",
	})
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{
		service: service,
	}
}
