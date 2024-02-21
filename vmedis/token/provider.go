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

type Provider struct {
	db             *Database
	reloadInterval time.Duration

	activeTokens     []string
	activeTokensLock sync.RWMutex

	closeCh   chan struct{}
	closeOnce sync.Once
}

func (m *Provider) GetActiveToken() (string, error) {
	m.activeTokensLock.RLock()
	defer m.activeTokensLock.RUnlock()

	if len(m.activeTokens) == 0 {
		return "", errors.New("no active tokens")
	}

	return m.activeTokens[rand.Intn(len(m.activeTokens))], nil
}

func (m *Provider) startReloader() {
	interval := max(m.reloadInterval, time.Minute)

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := m.ReloadTokens(context.Background()); err != nil {
				log.Printf("Error refreshing tokens: %s", err)
			}

		case <-m.closeCh:
			ticker.Stop()
			return
		}
	}
}

func (m *Provider) ReloadTokens(ctx context.Context) error {
	log.Println("Reloading tokens")

	activeTokens, err := m.db.GetNonExpiredTokens(ctx)
	if err != nil {
		return fmt.Errorf("get active tokens from DB: %w", err)
	}

	activeTokenStrings := slices2.Map(activeTokens, func(t models.VmedisToken) string {
		return t.Token
	})

	m.activeTokensLock.Lock()
	m.activeTokens = activeTokenStrings
	m.activeTokensLock.Unlock()

	log.Printf("Finished reloading, got %d active tokens", len(activeTokens))

	return nil
}

func (m *Provider) Close() {
	m.closeOnce.Do(func() {
		close(m.closeCh)
	})
}

func NewProvider(db *gorm.DB, reloadInterval time.Duration) (*Provider, error) {
	provider := &Provider{
		db:             NewDatabase(db),
		reloadInterval: reloadInterval,
		closeCh:        make(chan struct{}),
	}

	if err := provider.ReloadTokens(context.Background()); err != nil {
		return nil, fmt.Errorf("initialize tokens: %w", err)
	}

	if len(provider.activeTokens) == 0 {
		return nil, errors.New("no active tokens")
	}

	go provider.startReloader()

	return provider, nil
}
