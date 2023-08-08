package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

// OutOfStockDrugsResponse is the response of the Out-of-Stock Drugs client method.
type OutOfStockDrugsResponse struct {
	Drugs      []DrugStock
	OtherPages []int
}

// DrugStock is the stock of a drug.
type DrugStock struct {
	Drug  Drug  `data-column:"<self>"`
	Stock Stock `data-column:"8"`
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisCode   string `data-column:"4"`
	Name         string `data-column:"5"`
	Manufacturer string `data-column:"12"`
	Supplier     string `data-column:"13"`
	MinimumStock Stock  `data-column:"6"`
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string
	Quantity float64
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (s *Stock) UnmarshalDataColumn(selection *goquery.Selection) error {
	stockString := selection.Text()
	stockString = strings.TrimSpace(stockString)

	split := strings.Split(stockString, " ")
	if len(split) > 0 {
		// Here, the string can be either in "1.000,00" format or "1000.00" format.
		// We need to predict which format it is and convert it to float64.

		var qStr string

		// If there are always 3 digits after the dot, then it is in "1.000,00" format. Otherwise, it is in "1000.00" format.
		if strings.Count(split[0], ".") == 0 {
			qStr = strings.ReplaceAll(split[0], ",", ".")
		} else {
			firstFormat := true

			dotSplit := strings.Split(split[0], ".")
			for i := 1; i < len(dotSplit); i++ {
				beforeComma := strings.Split(dotSplit[i], ",")[0]
				if len(beforeComma) != 3 {
					firstFormat = false
					break
				}
			}

			if firstFormat {
				qStr = strings.ReplaceAll(split[0], ".", "")
				qStr = strings.ReplaceAll(qStr, ",", ".")
			} else {
				qStr = split[0]
			}
		}

		q, err := strconv.ParseFloat(qStr, 64)
		if err != nil {
			return fmt.Errorf("parse quantity from string [%s]: %w", split[0], err)
		}

		s.Quantity = q
	}

	if len(split) > 1 {
		s.Unit = split[1]
	}

	return nil
}
