package token

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
)

type Refresher struct {
	db              *Database
	refresher       ExternalRefresher
	refreshInterval time.Duration
}

func (m *Refresher) RefreshTokens(ctx context.Context) error {
	return m.db.Transaction(ctx, func(tx *Database) error {
		nonExpiredTokens, err := tx.GetNonExpiredTokens(ctx)
		if err != nil {
			return fmt.Errorf("get non expired tokens from DB: %w", err)
		}

		log.Printf("Refreshing %d tokens", len(nonExpiredTokens))

		refreshResult, err := m.refresher.RefreshTokens(ctx, nonExpiredTokens)
		if err != nil {
			return fmt.Errorf("refresh tokens: %w", err)
		}

		updatedTokens := make([]models.VmedisToken, 0, len(refreshResult))
		for token, status := range refreshResult {
			updatedTokens = append(updatedTokens, models.VmedisToken{
				Token:   token,
				State:   status.State,
				Version: status.Version,
			})
		}

		if err := tx.UpsertTokensState(ctx, updatedTokens); err != nil {
			return fmt.Errorf("upsert tokens state: %w", err)
		}

		activeTokens := slices2.Filter(nonExpiredTokens, func(token models.VmedisToken) bool {
			return refreshResult[token.Token].State == models.TokenStateActive
		})

		log.Printf("Finished refreshing, got %d active tokens", len(activeTokens))

		return nil
	})
}

func NewRefresher(db *gorm.DB, externalRefresher ExternalRefresher) *Refresher {
	refresher := &Refresher{
		db:        NewDatabase(db),
		refresher: externalRefresher,
	}

	return refresher
}
