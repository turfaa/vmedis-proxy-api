package proxy

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// SaleStatisticsResponse represents the sale statistics API response.
type SaleStatisticsResponse struct {
	History []SaleStatistics `json:"history"`
}

// SaleStatistics represents the sale statistics from the beginning of the day until PulledAt.
type SaleStatistics struct {
	PulledAt      time.Time `json:"pulledAt"`
	TotalSales    float64   `json:"totalSales"`
	NumberOfSales int       `json:"numberOfSales"`
}

// FromSalesStatisticsClientSchema converts SaleStatistics from client schema to proxy schema.
func FromSalesStatisticsClientSchema(pulledAt time.Time, salesStatistics client.SalesStatistics) (SaleStatistics, error) {
	totalSales, err := salesStatistics.TotalSalesFloat64()
	if err != nil {
		return SaleStatistics{}, fmt.Errorf("total sales float64: %w", err)
	}

	return SaleStatistics{
		PulledAt:      pulledAt,
		TotalSales:    totalSales,
		NumberOfSales: salesStatistics.NumberOfSales,
	}, nil
}

// FromModelsSaleStatistics converts SaleStatistics from models.SaleStatistics to proxy schema.
func FromModelsSaleStatistics(saleStatistics models.SaleStatistics) SaleStatistics {
	return SaleStatistics{
		PulledAt:      saleStatistics.PulledAt,
		TotalSales:    saleStatistics.TotalSales,
		NumberOfSales: saleStatistics.NumberOfSales,
	}
}

// HandleGetDailySalesStatistics handles the request to get the daily sales statistics.
func (s *ApiServer) HandleGetDailySalesStatistics(c *gin.Context) {
	var modelStats []models.SaleStatistics
	if err := s.DB.
		Where("pulled_at >= ?", beginningOfToday().Add(30*time.Minute)).
		Order("pulled_at ASC").
		Find(&modelStats).
		Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get historical statistics from DB: " + err.Error(),
		})
		return
	}

	latestStat, err := s.Client.GetDailySalesStatistics(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get latest statistics from API: " + err.Error(),
		})
		return
	}

	var stats []SaleStatistics
	for _, s := range modelStats {
		stats = append(stats, FromModelsSaleStatistics(s))
	}

	latest, err := FromSalesStatisticsClientSchema(time.Now(), latestStat)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to convert latest statistics from API: " + err.Error(),
		})
		return
	}

	if len(stats) == 0 || (stats[len(stats)-1].TotalSales <= latest.TotalSales && stats[len(stats)-1].PulledAt.Before(latest.PulledAt)) {
		stats = append(stats, latest)
	}

	c.JSON(200, SaleStatisticsResponse{History: stats})
}

func beginningOfToday() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}
