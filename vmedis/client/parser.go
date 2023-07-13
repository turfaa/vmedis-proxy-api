package client

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
)

var (
	numberOfSalesRegexp = regexp.MustCompile("data dari total ([0-9]+) data")
	totalSalesRegexp    = regexp.MustCompile("Total Penjualan : ([0-9.,]+)")
)

// ParseSalesStatistics parses the sales statistics from the given reader.
func ParseSalesStatistics(r io.Reader) (SalesStatistics, error) {
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error reading response body: %w", err)
	}

	body := string(bodyBytes)

	numberOfSalesMatches := numberOfSalesRegexp.FindStringSubmatch(body)
	if len(numberOfSalesMatches) != 2 {
		return SalesStatistics{}, fmt.Errorf("error parsing number of sales, matches: %v", numberOfSalesMatches)
	}

	numberOfSales, err := strconv.Atoi(numberOfSalesMatches[1])
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error parsing number of sales [%s]: %w", numberOfSalesMatches[1], err)
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
