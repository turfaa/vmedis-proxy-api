package token

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/slices2"
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

		nonExpiredTokenStrings := slices2.Map(nonExpiredTokens, func(t models.VmedisToken) string {
			return t.Token
		})

		log.Printf("Refresing %d tokens", len(nonExpiredTokenStrings))

		refreshResult, err := m.refresher.RefreshTokens(ctx, nonExpiredTokenStrings)
		if err != nil {
			return fmt.Errorf("refresh tokens: %w", err)
		}

		updatedTokens := make([]models.VmedisToken, 0, len(refreshResult))
		for token, state := range refreshResult {
			updatedTokens = append(updatedTokens, models.VmedisToken{
				Token: token,
				State: state,
			})
		}

		if err := tx.UpsertTokensState(ctx, updatedTokens); err != nil {
			return fmt.Errorf("upsert tokens state: %w", err)
		}

		activeTokens := slices2.Filter(nonExpiredTokenStrings, func(token string) bool {
			return refreshResult[token] == models.TokenStateActive
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
