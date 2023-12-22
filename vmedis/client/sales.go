package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// SalesResponse is the response of Sales client method.
type SalesResponse struct {
	Sales      []Sale
	OtherPages []int
}

// GetAllTodaySales gets all sales from today from vmedis.
// It starts with getting the number of pages by calling the API with page 9999. The last page is the number of pages.
// Then it calls the /apt-lap-penjualanobat-batch/index?page=<page> page and try to parse the sales from it.
func (c *Client) GetAllTodaySales(ctx context.Context) ([]Sale, error) {
	var (
		sales []Sale
		pages = make(chan int, c.concurrency*2)
		wg    sync.WaitGroup
		lock  sync.Mutex
	)

	// Get the number of pages
	log.Println("Getting number of pages of today sales")
	res, err := c.GetTodaySales(ctx, 9999)
	if err != nil {
		return nil, fmt.Errorf("get number of pages: %w", err)
	}

	lastPage := 1
	for _, p := range res.OtherPages {
		if p > lastPage {
			lastPage = p
		}
	}

	log.Printf("Number of today sales pages: %d\n", lastPage)

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
				log.Printf("Getting today sales at page %d\n", page)

				res, err := c.GetTodaySales(ctx, page)
				if err != nil {
					log.Printf("Error getting today sales at page #%d: %v\n", page, err)
					continue
				}

				lock.Lock()
				sales = append(sales, res.Sales...)
				lock.Unlock()

				log.Printf("Got %d sales at page %d\n", len(res.Sales), page)
			}
		}()
	}

	wg.Wait()

	return sales, nil
}

// GetTodaySales gets one page of the sales from today from vmedis.
// It calls the /apt-lap-penjualanobat-batch/index?page=<page> page and try to parse the sales from it.
func (c *Client) GetTodaySales(ctx context.Context, page int) (SalesResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/apt-lap-penjualanobat-batch/index?page=%d", page))
	if err != nil {
		return SalesResponse{}, fmt.Errorf("get today sales at page #%d: %w", page, err)
	}
	defer res.Body.Close()

	sales, err := ParseSales(res.Body)
	if err != nil {
		return SalesResponse{}, fmt.Errorf("parse today sales at page #%d: %w", page, err)
	}

	return sales, nil
}

// ParseSales parses the sales from the given reader.
// It usually comes from the /apt-lap-penjualanobat-batch/index?page=<page> page.
func ParseSales(r io.Reader) (SalesResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return SalesResponse{}, fmt.Errorf("new document from reader: %w", err)
	}

	var sales []Sale
	doc.Find("tr[data-key]").Each(func(i int, s *goquery.Selection) {
		sale, err := parseSale(s)
		if err != nil {
			log.Printf("error parsing sale #%d: %s", i, err)
			return
		}

		sales = append(sales, sale)
	})

	return SalesResponse{Sales: sales, OtherPages: parsePagination(doc)}, nil
}

func parseSale(selection *goquery.Selection) (Sale, error) {
	var sale Sale
	if err := UnmarshalDataColumn("sales-column", selection, &sale); err != nil {
		return Sale{}, fmt.Errorf("unmarshal sale: %w", err)
	}

	selection.Find("table tr:nth-child(n+2)").Each(func(i int, s *goquery.Selection) {
		su, err := parseSaleUnit(s)
		if err != nil {
			log.Printf("error parsing sale unit #%d: %s", i, err)
			return
		}

		sale.SaleUnits = append(sale.SaleUnits, su)
	})

	// Get the value from <button type="button" class="btn btn-warning btn-xs actionPrint" value="110844" title="Cetak Faktur">.
	idStr, ok := selection.Find("button.actionPrint").Attr("value")
	if ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return Sale{}, fmt.Errorf("parse sale vmedis id: %w", err)
		}

		sale.ID = id
	} else {
		html, _ := selection.Html()
		return Sale{}, fmt.Errorf("sale vmedis id not found in: %s", html)
	}

	return sale, nil
}

func parseSaleUnit(selection *goquery.Selection) (SaleUnit, error) {
	var su SaleUnit
	if err := UnmarshalDataColumnByIndex("sales-index", selection, &su); err != nil {
		return SaleUnit{}, fmt.Errorf("unmarshal sale unit: %w", err)
	}

	return su, nil
}
