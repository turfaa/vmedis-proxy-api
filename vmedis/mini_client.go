package vmedis

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// MiniClient is created to place methods that can't be placed in the main Client.
type MiniClient struct {
	baseUrl    string
	httpClient *http.Client
	limiter    *rate.Limiter
}

// RefreshTokens can't be placed in the main Client
// because it is needed by the token manager.
// We probably need to rethink the code/dependency design.
func (c *MiniClient) RefreshTokens(ctx context.Context, tokens []string) (map[string]models.TokenState, error) {
	result := make(map[string]models.TokenState, len(tokens))
	lock := sync.Mutex{}

	var errs errgroup.Group

	for _, t := range tokens {
		token := t
		errs.Go(func() error {
			res, err := c.getWithSessionId(ctx, "/", token)
			if err != nil {
				return fmt.Errorf("error refreshing token: %w", err)
			}
			defer res.Body.Close()

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %w", err)
			}

			body := string(bodyBytes)
			if strings.Contains(body, "Vmedis - Login") {
				lock.Lock()
				result[token] = models.TokenStateExpired
				lock.Unlock()
				return nil
			}

			if !strings.Contains(body, "Vmedis - Beranda") {
				return fmt.Errorf("unknown response body: %s", body)
			}

			lock.Lock()
			result[token] = models.TokenStateActive
			lock.Unlock()

			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *MiniClient) getWithSessionId(ctx context.Context, path, sessionId string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("error waiting for limiter: %w", err)
	}

	finalPath := c.baseUrl + path
	log.Printf("GET %s", finalPath)

	req, err := http.NewRequest("GET", finalPath, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Cookie", "PHPSESSID="+sessionId)
	req = req.WithContext(ctx)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request with session id %s: %w", sessionId, err)
	}

	return res, nil
}

func NewMini(baseUrl string, limiter *rate.Limiter) *MiniClient {
	return &MiniClient{
		baseUrl:    baseUrl,
		httpClient: &http.Client{},
		limiter:    limiter,
	}
}
