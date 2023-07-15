package client

import (
	"strconv"
	"strings"
)

// SalesStatistics is the statistics of sales in a period of time.
type SalesStatistics struct {
	// NumberOfSales is the number of sales in the period of time.
	NumberOfSales int

	// TotalSales is the total amount of sales in the period of time.
	// This is in IDR.
	// For precision purposes, this is still represented as string.
	TotalSales string
}

// TotalSalesFloat64 returns the TotalSales as float64.
func (s SalesStatistics) TotalSalesFloat64() (float64, error) {
	ts := strings.ReplaceAll(s.TotalSales, ".", "")
	ts = strings.ReplaceAll(ts, ",", ".")
	return strconv.ParseFloat(ts, 64)
}
