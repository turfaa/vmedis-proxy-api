package vmedis

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"

	"github.com/PuerkitoBio/goquery"
)

// GetDrug gets the drug details from vmedis.
// It calls the /obat-batch/view?id=<id> page and try to parse the drug from it.
func (c *Client) GetDrug(ctx context.Context, id int) (Drug, error) {
	res, err := c.get(ctx, fmt.Sprintf("/obat-batch/view?id=%d", id))
	if err != nil {
		return Drug{}, fmt.Errorf("get drug: %w", err)
	}
	defer res.Body.Close()

	drug, err := ParseDrugDetails(res.Body)
	if err != nil {
		return Drug{}, fmt.Errorf("parse drug: %w", err)
	}

	drug.VmedisID = id
	return drug, nil
}

// ParseDrugDetails parses the drug from the given reader
// It usually comes from the /obat-batch/view?id=<id> page.
func ParseDrugDetails(r io.Reader) (Drug, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return Drug{}, fmt.Errorf("parse HTML: %w", err)
	}

	var drug Drug
	if err := UnmarshalForm(doc.Selection, &drug); err != nil {
		return Drug{}, fmt.Errorf("unmarshal drug details: %w", err)
	}

	units, err := EnrichUnitsFromDoc(drug.Units, doc)
	if err != nil {
		return Drug{}, fmt.Errorf("enrich units of drug %s: %w", drug.Name, err)
	}

	drug.Units = units
	if len(units) > 0 {
		drug.MinimumStock.Unit = units[0].Unit
	}

	stocks, err := ParseStocksInDrugDetails(doc)
	if err != nil {
		return Drug{}, fmt.Errorf("parse stocks in drug details: %w", err)
	}

	sortStocksByUnit(stocks, units)
	drug.Stocks = stocks
	return drug, nil
}

// ParseStocksInDrugDetails parses the stocks in the drug details page.
func ParseStocksInDrugDetails(doc *goquery.Document) ([]Stock, error) {
	type shadowStock struct {
		Unit     string  `column-index:"4"`
		Quantity float64 `column-index:"3"`
	}

	stocksTab := doc.Find("div#detail")

	var stocks []Stock
	stocksTab.Find("tr[data-key]").Each(func(i int, s *goquery.Selection) {
		var stock shadowStock
		if err := UnmarshalDataColumnByIndex("column-index", s, &stock); err != nil {
			log.Printf("error parsing stock #%d: %s", i, err)
			return
		}

		stocks = append(stocks, Stock{
			Unit:     stock.Unit,
			Quantity: stock.Quantity,
		})
	})

	return compactStocks(stocks), nil
}

func compactStocks(stocks []Stock) []Stock {
	stockByUnit := make(map[string]Stock)

	for _, stock := range stocks {
		if stock.Quantity == 0 {
			continue
		}

		current, ok := stockByUnit[stock.Unit]
		if !ok {
			stockByUnit[stock.Unit] = stock
			continue
		}

		current.Quantity += stock.Quantity
		stockByUnit[stock.Unit] = current
	}

	compact := make([]Stock, 0, len(stockByUnit))
	for _, stock := range stockByUnit {
		compact = append(compact, stock)
	}

	return compact
}

func sortStocksByUnit(stocks []Stock, units []Unit) {
	unitOrder := make(map[string]int, len(units))
	for _, unit := range units {
		unitOrder[unit.Unit] = unit.UnitOrder
	}

	sort.Slice(stocks, func(i, j int) bool {
		return unitOrder[stocks[i].Unit] > unitOrder[stocks[j].Unit]
	})
}
