package vmedisv2

import (
	"context"
	"fmt"
	"time"
)

// SalesStatistics is the statistics of sales in a period of time.
type SalesStatistics struct {
	// NumberOfSales is the number of sales in the period of time.
	NumberOfSales int

	// TotalSales is the total amount of sales in the period of time, in IDR.
	TotalSales float64
}

// GetSalesStatistics gets the statistics of the sales between the given
// dates, inclusive. Only the date parts are used.
// It calls the endpoint the "Lap Penjualan Obat" v2 report page
// (/apt-lap-penjualanobat-batch-v2) gets its totals from.
func (c *Client) GetSalesStatistics(ctx context.Context, from, to time.Time) (SalesStatistics, error) {
	var res struct {
		Total    *int     `json:"total"`
		TotalAll *float64 `json:"totalall"`

		// Set instead of the fields above when the request is rejected,
		// e.g. {"success":false,"message":"appId is required"}.
		Message string `json:"message"`
	}

	err := c.callReport(ctx, "apt-lap-penjualanobat-batch/indexcount-v2", map[string]any{
		"tanggalawal":  from.Format(time.DateOnly),
		"tanggalakhir": to.Format(time.DateOnly),
	}, &res)
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("get sales statistics between %s and %s: %w", from.Format(time.DateOnly), to.Format(time.DateOnly), err)
	}

	if res.Total == nil || res.TotalAll == nil {
		return SalesStatistics{}, fmt.Errorf("get sales statistics between %s and %s: unexpected response (message: %q)", from.Format(time.DateOnly), to.Format(time.DateOnly), res.Message)
	}

	return SalesStatistics{
		NumberOfSales: *res.Total,
		TotalSales:    *res.TotalAll,
	}, nil
}
