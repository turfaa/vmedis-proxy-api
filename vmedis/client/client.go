package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Client is the main struct for the vmedis client.
type Client struct {
	BaseUrl    string
	SessionIds []string

	httpClient  *http.Client
	concurrency int
	limiter     *rate.Limiter
}

// New creates a new client.
func New(baseUrl string, sessionIds []string, concurrency int, limiter *rate.Limiter) *Client {
	return &Client{
		BaseUrl:     baseUrl,
		SessionIds:  sessionIds,
		httpClient:  &http.Client{},
		concurrency: concurrency,
		limiter:     limiter,
	}
}

// AutoRefreshSessionIds refreshes the session ids of the client every d duration.
// It returns a function to stop the auto refresh.
func (c *Client) AutoRefreshSessionIds(d time.Duration) func() {
	ticker := time.NewTicker(d)

	stop := make(chan struct{})
	go func() {
		for {
			log.Println("Refreshing session ids")
			if err := c.RefreshSessionIds(context.Background()); err != nil {
				log.Printf("Error refreshing session ids: %v\n", err)
			} else {
				log.Println("Session ids refreshed")
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

// RefreshSessionIds refreshes the session ids of the client.
// This is used to keep the sessions alive. We are doing this by calling the home page.
func (c *Client) RefreshSessionIds(ctx context.Context) error {
	var errs errgroup.Group

	for _, s := range c.SessionIds {
		sessionId := s
		errs.Go(func() error {
			res, err := c.getWithSessionId(ctx, "/", sessionId)
			if err != nil {
				return fmt.Errorf("error refreshing session id: %w", err)
			}
			defer res.Body.Close()

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
		})
	}

	return errs.Wait()
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	return c.getWithSessionId(ctx, path, c.SessionIds[rnd.Intn(len(c.SessionIds))])
}

func (c *Client) getWithSessionId(ctx context.Context, path, sessionId string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseUrl+path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Cookie", "PHPSESSID="+sessionId)
	req = req.WithContext(ctx)

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("error waiting for limiter: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	return res, nil
}
