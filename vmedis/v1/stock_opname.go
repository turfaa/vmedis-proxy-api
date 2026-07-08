package vmedisv1

import (
	"context"
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

// StockOpnamesResponse is the response of StockOpnames client method.
type StockOpnamesResponse struct {
	StockOpnames []StockOpname
	OtherPages   []int
}

// GetAllTodayStockOpnames gets all stock opnames from today from vmedis.
// It fetches every /laporan-stokopname-batch/index?page=<page> page
// concurrently and returns an error if any page cannot be fetched or parsed.
func (c *Client) GetAllTodayStockOpnames(ctx context.Context) ([]StockOpname, error) {
	stockOpnames, err := getAllPages(ctx, "today stock opnames", c.concurrency, func(ctx context.Context, page int) ([]StockOpname, []int, error) {
		res, err := c.GetTodayStockOpnames(ctx, page)
		if err != nil {
			return nil, nil, err
		}

		return res.StockOpnames, res.OtherPages, nil
	})
	if err != nil {
		return nil, err
	}

	augmentDuplicatedStockOpnameCodes(stockOpnames)
	return stockOpnames, nil
}

// GetTodayStockOpnames gets all stock opnames from today from vmedis.
// It calls the /laporan-stokopname-batch/index?page=<page> page and try to parse the stock opnames from it.
func (c *Client) GetTodayStockOpnames(ctx context.Context, page int) (StockOpnamesResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/laporan-stokopname-batch/index?page=%d", page))
	if err != nil {
		return StockOpnamesResponse{}, fmt.Errorf("get stock opnames: %w", err)
	}
	defer res.Body.Close()

	sos, err := ParseStockOpnames(res.Body)
	if err != nil {
		return StockOpnamesResponse{}, fmt.Errorf("parse stock opnames: %w", err)
	}

	augmentDuplicatedStockOpnameCodes(sos.StockOpnames)
	return sos, nil
}

// ParseStockOpnames parses the stock opnames from the API response.
func ParseStockOpnames(r io.Reader) (StockOpnamesResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return StockOpnamesResponse{}, fmt.Errorf("create goquery document from reader: %w", err)
	}

	var stockOpnames []StockOpname
	doc.Find("tr[data-key]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		so, parseErr := parseStockOpname(s)
		if parseErr != nil {
			err = fmt.Errorf("parse stock opname #%d: %w", i, parseErr)
			return false
		}

		stockOpnames = append(stockOpnames, so)
		return true
	})
	if err != nil {
		return StockOpnamesResponse{}, err
	}

	return StockOpnamesResponse{StockOpnames: stockOpnames, OtherPages: parsePagination(doc)}, nil
}

func parseStockOpname(selection *goquery.Selection) (StockOpname, error) {
	var so StockOpname
	if err := UnmarshalDataColumnByIndex("so-index", selection, &so); err != nil {
		return StockOpname{}, err
	}

	return so, nil
}

func augmentDuplicatedStockOpnameCodes(stockOpnames []StockOpname) {
	visitedID := make(map[string]bool, len(stockOpnames))

	for i := range stockOpnames {
		if visitedID[stockOpnames[i].ID] {
			for x := 2; ; x++ {
				id := fmt.Sprintf("%s-augmented-because-duplicate-%d", stockOpnames[i].ID, x)
				if !visitedID[id] {
					stockOpnames[i].ID = id
					break
				}
			}
		}

		visitedID[stockOpnames[i].ID] = true
	}
}
