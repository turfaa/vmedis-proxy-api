package vmedis

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/turfaa/vmedis-proxy-api/vmedis/token"
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Client is the main struct for the vmedis client.
type Client struct {
	BaseUrl string

	httpClient  *http.Client
	concurrency int
	limiter     *rate.Limiter

	tokenManager *token.Manager
}

// New creates a new client.
func New(
	baseUrl string,
	concurrency int,
	limiter *rate.Limiter,
	tokenManager *token.Manager,
) *Client {
	return &Client{
		BaseUrl:      baseUrl,
		httpClient:   &http.Client{},
		concurrency:  concurrency,
		limiter:      limiter,
		tokenManager: tokenManager,
	}
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	sessionId, err := c.tokenManager.GetActiveToken()
	if err != nil {
		return nil, fmt.Errorf("get active session id: %w", err)
	}

	return c.getWithSessionId(ctx, path, sessionId)
}

func (c *Client) getWithSessionId(ctx context.Context, path, sessionId string) (*http.Response, error) {
	finalPath := c.BaseUrl + path
	log.Printf("GET %s\n", finalPath)

	req, err := http.NewRequest("GET", finalPath, nil)
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
		return nil, fmt.Errorf("error executing request with session id %s: %w", sessionId, err)
	}

	return res, nil
}
