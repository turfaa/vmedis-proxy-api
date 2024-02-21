package vmedis

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

func (c *Client) RefreshTokens(ctx context.Context, tokens []string) (map[string]models.TokenState, error) {
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
