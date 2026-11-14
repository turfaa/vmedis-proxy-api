package vmedisv1

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// salesPagePaths are the sales report pages of each vmedis version. A session
// can open only the page of its own version; vmedis redirects it from the
// other one.
var salesPagePaths = map[models.TokenVersion]string{
	models.TokenVersionV1: "/apt-lap-penjualanobat-batch/index",
	models.TokenVersionV2: "/apt-lap-penjualanobat-batch-v2",
}

// RefreshTokens checks every token against vmedis. A token is expired when
// vmedis responds with the login page. Otherwise its version is detected
// from which sales page it can open: the v1 one or the v2 one. A token that
// can open neither is useless to every client, so it is marked expired too.
func (c *Client) RefreshTokens(ctx context.Context, tokens []models.VmedisToken) (map[string]models.TokenStatus, error) {
	result := make(map[string]models.TokenStatus, len(tokens))
	lock := sync.Mutex{}

	var errs errgroup.Group

	for _, token := range tokens {
		errs.Go(func() error {
			status, err := c.checkToken(ctx, token)
			if err != nil {
				return fmt.Errorf("check token %d: %w", token.ID, err)
			}

			lock.Lock()
			result[token.Token] = status
			lock.Unlock()

			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// checkToken opens the sales page of each version with the token, starting
// with the token's last known version, until one opens.
func (c *Client) checkToken(ctx context.Context, token models.VmedisToken) (models.TokenStatus, error) {
	versions := []models.TokenVersion{models.TokenVersionV1, models.TokenVersionV2}
	if token.Version == models.TokenVersionV2 {
		versions = []models.TokenVersion{models.TokenVersionV2, models.TokenVersionV1}
	}

	expired := models.TokenStatus{State: models.TokenStateExpired, Version: token.Version}

	for _, version := range versions {
		res, err := c.getWithSessionId(ctx, salesPagePaths[version], token.Token)
		switch {
		case errors.Is(err, ErrInvalidToken):
			return expired, nil

		case errors.Is(err, ErrRedirected):
			continue

		case err != nil:
			return models.TokenStatus{}, fmt.Errorf("open %s sales page: %w", version, err)
		}

		res.Body.Close()
		return models.TokenStatus{State: models.TokenStateActive, Version: version}, nil
	}

	return expired, nil
}
