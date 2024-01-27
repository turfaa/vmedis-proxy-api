package proxy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/proxy/schema"
	"github.com/turfaa/vmedis-proxy-api/time2"
)

// HandleGetSalesStatistics handles the request to get the sales statistics.
func (s *ApiServer) HandleGetSalesStatistics(c *gin.Context) {
	from, until, err := time2.GetTimeRangeFromQuery(c)
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

	dailyHistory := generateDailyHistory(stats)
	c.JSON(200, schema.SaleStatisticsResponse{History: stats, DailyHistory: dailyHistory})
}

// HandleGetDailySalesStatistics handles the request to get the daily sales statistics.
func (s *ApiServer) HandleGetDailySalesStatistics(c *gin.Context) {
	from, until := time2.Today()

	stats, err := s.getSalesStatistics(c, from, until)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get historical statistics: " + err.Error(),
		})
		return
	}

	dailyHistory := generateDailyHistory(stats)
	c.JSON(200, schema.SaleStatisticsResponse{History: stats, DailyHistory: dailyHistory})
}

func (s *ApiServer) getSalesStatistics(ctx context.Context, from, until time.Time) ([]schema.SaleStatistics, error) {
	var modelStats []models.SaleStatistics
	if err := s.db.
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

	if until == time2.EndOfToday() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if latestStat, err := s.client.GetDailySalesStatistics(ctx); err != nil {
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

func generateDailyHistory(stats []schema.SaleStatistics) []schema.SaleStatistics {
	res := make([]schema.SaleStatistics, 0, len(stats))
	for _, stat := range stats {
		if len(res) == 0 || !statsInSameDay(res[len(res)-1], stat) {
			res = append(res, stat)
		} else {
			res[len(res)-1] = stat
		}
	}

	return res
}

func statsInSameDay(stat1, stat2 schema.SaleStatistics) bool {
	return stat1.PulledAt.Year() == stat2.PulledAt.Year() &&
		stat1.PulledAt.Month() == stat2.PulledAt.Month() &&
		stat1.PulledAt.Day() == stat2.PulledAt.Day()
}
