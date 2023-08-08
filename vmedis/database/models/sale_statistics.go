package models

import (
	"time"
)

// SaleStatistics represents the sale statistics from the beginning of the day until PulledAt.
type SaleStatistics struct {
	ID            uint      `gorm:"primarykey"`
	PulledAt      time.Time `gorm:"index"`
	TotalSales    float64
	NumberOfSales int
}
