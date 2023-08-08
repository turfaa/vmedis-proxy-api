package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the main struct for the vmedis client.
type Client struct {
	BaseUrl   string
	SessionId string

	httpClient  *http.Client
	concurrency int
}

// New creates a new client.
func New(baseUrl, sessionId string, concurrency int) *Client {
	return &Client{
		BaseUrl:     baseUrl,
		SessionId:   sessionId,
		httpClient:  &http.Client{},
		concurrency: concurrency,
	}
}

// AutoRefreshSessionId refreshes the session id of the client every d duration.
// It returns a function to stop the auto refresh.
func (c *Client) AutoRefreshSessionId(d time.Duration) func() {
	ticker := time.NewTicker(d)

	stop := make(chan struct{})
	go func() {
		for {
			log.Println("Refreshing session id")
			if err := c.RefreshSessionId(context.Background()); err != nil {
				log.Printf("Error refreshing session id: %v\n", err)
			} else {
				log.Println("Session id refreshed")
			}

			select {
			case <-ticker.C:

			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(stop)
	}
}

// RefreshSessionId refreshes the session id of the client.
// This is used to keep the session alive. We are doing this by calling the home page.
func (c *Client) RefreshSessionId(ctx context.Context) error {
	res, err := c.get(ctx, "/")
	if err != nil {
		return fmt.Errorf("error refreshing session id: %w", err)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	body := string(bodyBytes)
	if strings.Contains(body, "Vmedis - Login") {
		return errors.New("session id expired")
	} else if !strings.Contains(body, "Vmedis - Beranda") {
		return fmt.Errorf("unknown response body: %s", body)
	}

	return nil
}

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

	log.Printf("Number of pages: %d\n", lastPage)

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

	drugs, err := ParseOutOfStockDrugs(res.Body)
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("parse out of stock drugs at page %d: %w", page, err)
	}

	return drugs, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseUrl+path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Cookie", "PHPSESSID="+c.SessionId)
	req = req.WithContext(ctx)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	return res, nil
}
