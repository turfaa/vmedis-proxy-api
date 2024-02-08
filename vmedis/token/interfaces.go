package token

import (
	"context"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type Refresher interface {
	RefreshTokens(ctx context.Context, tokens []string) (map[string]models.TokenState, error)
}
