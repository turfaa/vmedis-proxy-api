package schema

import (
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// SaleStatisticsResponse represents the sale statistics API response.
type SaleStatisticsResponse struct {
	History      []SaleStatistics `json:"history"`
	DailyHistory []SaleStatistics `json:"dailyHistory"`
}

// SaleStatistics represents the sale statistics from the beginning of the day until PulledAt.
type SaleStatistics struct {
	PulledAt      time.Time `json:"pulledAt"`
	TotalSales    float64   `json:"totalSales"`
	NumberOfSales int       `json:"numberOfSales"`
}

// FromSalesStatisticsClientSchema converts SaleStatistics from client schema to proxy schema.
func FromSalesStatisticsClientSchema(pulledAt time.Time, salesStatistics vmedis.SalesStatistics) (SaleStatistics, error) {
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