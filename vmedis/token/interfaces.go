package token

import (
	"context"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type ExternalRefresher interface {
	RefreshTokens(ctx context.Context, tokens []string) (map[string]models.TokenState, error)
}
