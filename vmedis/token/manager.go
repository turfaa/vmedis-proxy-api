package token

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/slices2"
)

type Manager struct {
	db              *Database
	refresher       Refresher
	refreshInterval time.Duration

	activeTokens     []string
	activeTokensLock sync.RWMutex

	closeCh   chan struct{}
	closeOnce sync.Once
}

func (m *Manager) GetActiveToken() (string, error) {
	m.activeTokensLock.RLock()
	defer m.activeTokensLock.RUnlock()

	if len(m.activeTokens) == 0 {
		return "", errors.New("no active tokens")
	}

	return m.activeTokens[rand.Intn(len(m.activeTokens))], nil
}

func (m *Manager) startRefresher() {
	interval := max(m.refreshInterval, time.Minute)

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := m.RefreshTokens(context.Background()); err != nil {
				log.Printf("Error refreshing tokens: %s", err)
			}

		case <-m.closeCh:
			ticker.Stop()
			return
		}
	}
}

func (m *Manager) RefreshTokens(ctx context.Context) error {
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

		m.activeTokensLock.Lock()
		m.activeTokens = activeTokens
		m.activeTokensLock.Unlock()

		log.Printf("Finished refreshing, got %d active tokens", len(activeTokens))

		return nil
	})
}

func (m *Manager) Close() {
	m.closeOnce.Do(func() {
		close(m.closeCh)
	})
}

func NewManager(
	db *gorm.DB,
	refresher Refresher,
	refreshInterval time.Duration,
) (*Manager, error) {
	manager := &Manager{
		db:              NewDatabase(db),
		refresher:       refresher,
		refreshInterval: refreshInterval,
		closeCh:         make(chan struct{}),
	}

	if err := manager.RefreshTokens(context.Background()); err != nil {
		return nil, fmt.Errorf("initialize tokens: %w", err)
	}

	if len(manager.activeTokens) == 0 {
		return nil, errors.New("no active tokens")
	}

	go manager.startRefresher()

	return manager, nil
}
