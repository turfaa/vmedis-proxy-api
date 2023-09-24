package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// OutOfStockDrugsResponse is the response of the Out-of-Stock Drugs client method.
type OutOfStockDrugsResponse struct {
	Drugs      []DrugStock
	OtherPages []int
}

// GetAllOutOfStockDrugs gets all the out-of-stock drugs from vmedis.
// It starts with getting the number of pages by calling the API with page 9999. The last page is the number of pages.
// Then it calls the /obathabis-batch/index?page=<page> page and try to parse the out-of-stock drugs from it.
func (c *Client) GetAllOutOfStockDrugs(ctx context.Context) ([]DrugStock, error) {
	var (
		wg    sync.WaitGroup
		drugs []DrugStock
		lock  sync.Mutex
		pages = make(chan int, c.concurrency*2)
	)

	// Get the number of pages
	log.Println("Getting number of pages of OoS drugs")
	res, err := c.GetOutOfStockDrugs(ctx, 9999)
	if err != nil {
		return nil, fmt.Errorf("get number of pages: %w", err)
	}

	lastPage := 0
	for _, p := range res.OtherPages {
		if p > lastPage {
			lastPage = p
		}
	}

	log.Printf("Number of OOS pages: %d\n", lastPage)

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
				log.Printf("Getting out of stock drugs at page %d\n", page)

				res, err := c.GetOutOfStockDrugs(ctx, page)
				if err != nil {
					log.Printf("Error getting out of stock drugs at page %d: %v\n", page, err)
					continue
				}

				log.Printf("Got %d out of stock drugs at page %d\n", len(res.Drugs), page)

				lock.Lock()
				drugs = append(drugs, res.Drugs...)
				lock.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(drugs) > 0 {
		return drugs, nil
	} else {
		return nil, errors.New("no out of stock drugs found, check the logs for errors")
	}
}

// GetOutOfStockDrugs gets the out-of-stock drugs from vmedis.
// It calls the /obathabis-batch/index?page=<page> page and try to parse the out-of-stock drugs from it.
func (c *Client) GetOutOfStockDrugs(ctx context.Context, page int) (OutOfStockDrugsResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/obathabis-batch/index?page=%d", page))
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("get out of stock drugs at page %d: %w", page, err)
	}
	defer res.Body.Close()

	drugs, err := ParseOutOfStockDrugs(res.Body)
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("parse out of stock drugs at page %d: %w", page, err)
	}

	return drugs, nil
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

	return OutOfStockDrugsResponse{Drugs: drugs, OtherPages: parsePagination(doc)}, nil
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
