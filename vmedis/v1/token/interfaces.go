package token

import (
	"context"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type ExternalRefresher interface {
	RefreshTokens(ctx context.Context, tokens []models.VmedisToken) (map[string]models.TokenStatus, error)
}
