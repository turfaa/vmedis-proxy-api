package proxy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetSalesStatistics handles the request to get the sales statistics.
func (s *ApiServer) HandleGetSalesStatistics(c *gin.Context) {
	from, until, err := getTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse time range query: %s", err),
		})
		return
	}

	stats, err := s.getSalesStatistics(c, from, until)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get historical statistics: " + err.Error(),
		})
		return
	}

	c.JSON(200, schema.SaleStatisticsResponse{History: stats})
}

// HandleGetDailySalesStatistics handles the request to get the daily sales statistics.
func (s *ApiServer) HandleGetDailySalesStatistics(c *gin.Context) {
	from, until := today()

	stats, err := s.getSalesStatistics(c, from, until)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get historical statistics: " + err.Error(),
		})
		return
	}

	c.JSON(200, schema.SaleStatisticsResponse{History: stats})
}

func (s *ApiServer) getSalesStatistics(ctx context.Context, from, until time.Time) ([]schema.SaleStatistics, error) {
	var modelStats []models.SaleStatistics
	if err := s.DB.
		Where("pulled_at >= ? AND pulled_at <= ?", from, until).
		Order("pulled_at ASC").
		Find(&modelStats).
		Error; err != nil {
		return nil, fmt.Errorf("failed to get historical statistics from DB: %w", err)
	}

	var stats []schema.SaleStatistics
	for _, s := range modelStats {
		stats = append(stats, schema.FromModelsSaleStatistics(s))
	}

	if until == endOfToday() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if latestStat, err := s.Client.GetDailySalesStatistics(ctx); err != nil {
			log.Printf("failed to get latest statistics from API: %s", err)
		} else {
			latest, err := schema.FromSalesStatisticsClientSchema(time.Now(), latestStat)
			if err != nil {
				return nil, fmt.Errorf("failed to convert latest statistics from API: %w", err)
			}

			if len(stats) == 0 || (stats[len(stats)-1].TotalSales <= latest.TotalSales && stats[len(stats)-1].PulledAt.Before(latest.PulledAt)) {
				stats = append(stats, latest)
			}
		}
	}

	return stats, nil
}
