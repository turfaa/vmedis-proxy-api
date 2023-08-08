package client

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	numberOfSalesRegexp = regexp.MustCompile("data dari total ([0-9]+) data")
	totalSalesRegexp    = regexp.MustCompile("Total Penjualan : ([0-9.,]+)")
)

// GetDailySalesStatistics gets the daily sales statistics from vmedis.
// It calls the /apt-lap-penjualanobat-batch page and try to parse the statistics from it.
func (c *Client) GetDailySalesStatistics(ctx context.Context) (SalesStatistics, error) {
	res, err := c.get(ctx, "/apt-lap-penjualanobat-batch")
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error getting total daily sales: %w", err)
	}

	stats, err := ParseSalesStatistics(res.Body)
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error parsing total daily sales: %w", err)
	}

	return stats, nil
}

// ParseSalesStatistics parses the sales statistics from the given reader.
func ParseSalesStatistics(r io.Reader) (SalesStatistics, error) {
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error reading response body: %w", err)
	}

	body := string(bodyBytes)

	numberOfSalesMatches := numberOfSalesRegexp.FindStringSubmatch(body)

	var numberOfSales int
	if len(numberOfSalesMatches) == 2 {
		numberOfSales, err = strconv.Atoi(numberOfSalesMatches[1])
		if err != nil {
			return SalesStatistics{}, fmt.Errorf("error parsing number of sales [%s]: %w", numberOfSalesMatches[1], err)
		}
	}

	totalSalesMatches := totalSalesRegexp.FindStringSubmatch(body)
	if len(totalSalesMatches) != 2 {
		return SalesStatistics{}, fmt.Errorf("error parsing total sales, matches: %v", totalSalesMatches)
	}

	return SalesStatistics{
		NumberOfSales: numberOfSales,
		TotalSales:    totalSalesMatches[1],
	}, nil
}

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
