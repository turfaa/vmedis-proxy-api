package client

// SalesStatistics is the statistics of sales in a period of time.
type SalesStatistics struct {
	// NumberOfSales is the number of sales in the period of time.
	NumberOfSales int `json:"numberOfSales"`

	// TotalSales is the total amount of sales in the period of time.
	// This is in IDR.
	// For precision purposes, this is still represented as string.
	TotalSales string `json:"totalSales"`
}
