package client

import (
	"fmt"
	"log"
	"net/http"
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
			if err := c.RefreshSessionId(); err != nil {
				log.Printf("Error refreshing session id: %v\n", err)
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
	log.Println("Refreshing session id")

	if _, err := c.get("/"); err != nil {
		return fmt.Errorf("error refreshing session id: %w", err)
	}

	return nil
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
