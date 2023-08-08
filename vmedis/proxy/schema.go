package proxy

import (
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// SaleStatisticsResponse represents the sale statistics API response.
type SaleStatisticsResponse struct {
	History []SaleStatistics `json:"history"`
}

// SaleStatistics represents the sale statistics from the beginning of the day until PulledAt.
type SaleStatistics struct {
	PulledAt      time.Time
	TotalSales    float64
	NumberOfSales int
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
