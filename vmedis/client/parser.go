package client

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
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

// ParseOutOfStockDrugs parses the out-of-stock drugs from the given reader.
func ParseOutOfStockDrugs(r io.Reader) (OutOfStockDrugsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("parse HTML: %w", err)
	}

	var drugs []DrugStock
	doc.Find("tr[data-key]").Each(func(i int, s *goquery.Selection) {
		drug, err := parseOutOfStockDrug(s)
		if err != nil {
			log.Printf("error parsing out-of-stock drug #%d: %s", i, err)
			return
		}

		drugs = append(drugs, drug)
	})

	var otherPages []int
	doc.Find(".pagination li a").Each(func(i int, s *goquery.Selection) {
		page, err := strconv.Atoi(s.Text())
		if err != nil {
			// expected, ignore
			return
		}

		otherPages = append(otherPages, page)
	})

	return OutOfStockDrugsResponse{Drugs: drugs, OtherPages: otherPages}, nil
}

func parseOutOfStockDrug(doc *goquery.Selection) (DrugStock, error) {
	var ds DrugStock
	if err := UnmarshalDataColumn("oos-column", doc, &ds); err != nil {
		return DrugStock{}, fmt.Errorf("parse drug: %w", err)
	}

	if ds.Drug.MinimumStock.Unit == "" {
		ds.Drug.MinimumStock.Unit = ds.Stock.Unit
	}

	return ds, nil
}
