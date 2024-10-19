package vmedis

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Client is the main struct for the vmedis client.
type Client struct {
	BaseUrl string

	httpClient  *http.Client
	concurrency int
	limiter     *rate.Limiter

	tokenProvider tokenProvider
}

// New creates a new client.
func New(
	baseUrl string,
	concurrency int,
	limiter *rate.Limiter,
	tokenProvider tokenProvider,
) *Client {
	return &Client{
		BaseUrl:       baseUrl,
		httpClient:    &http.Client{Timeout: time.Minute},
		concurrency:   concurrency,
		limiter:       limiter,
		tokenProvider: tokenProvider,
	}
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	sessionId, err := c.tokenProvider.GetActiveToken()
	if err != nil {
		return nil, fmt.Errorf("get active session id: %w", err)
	}

	return c.getWithSessionId(ctx, path, sessionId)
}

func (c *Client) getWithSessionId(ctx context.Context, path, sessionId string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("error waiting for limiter: %w", err)
	}

	finalPath := c.BaseUrl + path
	log.Printf("GET %s", finalPath)

	req, err := http.NewRequest("GET", finalPath, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Cookie", "vmedisApp="+sessionId)
	req = req.WithContext(ctx)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request with session id %s: %w", sessionId, err)
	}

	return res, nil
}
