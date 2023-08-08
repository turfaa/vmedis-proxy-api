package models

import (
	"time"

	"gorm.io/gorm"
)

// SaleStatistics represents the sale statistics from the beginning of the day until PulledAt.
type SaleStatistics struct {
	gorm.Model
	PulledAt      time.Time
	TotalSales    float64
	NumberOfSales int
}
