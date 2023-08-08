package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Client is the main struct for the vmedis client.
type Client struct {
	BaseUrl   string
	SessionId string

	httpClient *http.Client
}

// New creates a new client.
func New(baseUrl, sessionId string) *Client {
	return &Client{
		BaseUrl:    baseUrl,
		SessionId:  sessionId,
		httpClient: &http.Client{},
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
			if err := c.RefreshSessionId(); err != nil {
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
func (c *Client) RefreshSessionId() error {
	res, err := c.get("/")
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
func (c *Client) GetDailySalesStatistics() (SalesStatistics, error) {
	res, err := c.get("/apt-lap-penjualanobat-batch")
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error getting total daily sales: %w", err)
	}

	stats, err := ParseSalesStatistics(res.Body)
	if err != nil {
		return SalesStatistics{}, fmt.Errorf("error parsing total daily sales: %w", err)
	}

	return stats, nil
}

func (c *Client) get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseUrl+path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Cookie", "PHPSESSID="+c.SessionId)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	return res, nil
}
