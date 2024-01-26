package vmedis

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// StockOpnamesResponse is the response of StockOpnames client method.
type StockOpnamesResponse struct {
	StockOpnames []StockOpname
	OtherPages   []int
}

// GetAllTodayStockOpnames gets all stock opnames from today from vmedis.
// It starts with getting the number of pages by calling the API with page 9999. The last page is the number of pages.
// Then it calls the /laporan-stokopname-batch/index?page=<page> page and try to parse the stock opnames from it.
func (c *Client) GetAllTodayStockOpnames(ctx context.Context) ([]StockOpname, error) {
	var (
		stockOpnames []StockOpname
		pages        = make(chan int, c.concurrency*2)
		wg           sync.WaitGroup
	)

	// Get the number of pages
	log.Println("Getting number of pages of today stock opnames")
	res, err := c.GetTodayStockOpnames(ctx, 9999)
	if err != nil {
		return nil, fmt.Errorf("get number of pages: %w", err)
	}

	lastPage := 1
	for _, p := range res.OtherPages {
		if p > lastPage {
			lastPage = p
		}
	}

	log.Printf("Number of today stock opnames pages: %d\n", lastPage)

	go func() {
		for i := 1; i <= lastPage; i++ {
			pages <- i
		}
		close(pages)
	}()

	// Start the workers
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for page := range pages {
				res, err := c.GetTodayStockOpnames(ctx, page)
				if err != nil {
					log.Printf("error getting today stock opnames page #%d: %s", page, err)
					continue
				}

				for _, so := range res.StockOpnames {
					stockOpnames = append(stockOpnames, so)
				}
			}
		}()
	}

	wg.Wait()
	return stockOpnames, nil
}

// GetTodayStockOpnames gets all stock opnames from today from vmedis.
// It calls the /laporan-stokopname-batch/index?page=<page> page and try to parse the stock opnames from it.
func (c *Client) GetTodayStockOpnames(ctx context.Context, page int) (StockOpnamesResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/laporan-stokopname-batch/index?page=%d", page))
	if err != nil {
		return StockOpnamesResponse{}, fmt.Errorf("get stock opnames: %w", err)
	}

	sos, err := ParseStockOpnames(res.Body)
	if err != nil {
		return StockOpnamesResponse{}, fmt.Errorf("parse stock opnames: %w", err)
	}

	return sos, nil
}

// ParseStockOpnames parses the stock opnames from the API response.
func ParseStockOpnames(r io.Reader) (StockOpnamesResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return StockOpnamesResponse{}, fmt.Errorf("create goquery document from reader: %w", err)
	}

	var stockOpnames []StockOpname
	doc.Find("tr[data-key]").Each(func(i int, s *goquery.Selection) {
		so, err := parseStockOpname(s)
		if err != nil {
			log.Printf("error parsing stock opname #%d: %s", i, err)
			return
		}

		stockOpnames = append(stockOpnames, so)
	})

	return StockOpnamesResponse{StockOpnames: stockOpnames, OtherPages: parsePagination(doc)}, nil
}

func parseStockOpname(selection *goquery.Selection) (StockOpname, error) {
	var so StockOpname
	if err := UnmarshalDataColumnByIndex("so-index", selection, &so); err != nil {
		return StockOpname{}, err
	}

	return so, nil
}
