package vmedisv1

import (
	"context"
	"errors"
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
			if errors.Is(err, ErrInvalidToken) {
				lock.Lock()
				result[token] = models.TokenStateExpired
				lock.Unlock()
				return nil
			}
			if err != nil {
				return fmt.Errorf("error refreshing token: %w", err)
			}
			defer res.Body.Close()

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %w", err)
			}

			if !strings.Contains(string(bodyBytes), "Aktifkan Menu V2") {
				return fmt.Errorf("unknown response body: %s", string(bodyBytes))
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
