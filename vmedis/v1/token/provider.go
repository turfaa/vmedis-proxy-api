package token

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
)

// Provider provides the non-expired tokens of one vmedis version.
type Provider struct {
	db             *Database
	version        models.TokenVersion
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
		return "", fmt.Errorf("no active %s tokens", m.version)
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
	log.Printf("Reloading %s tokens", m.version)

	activeTokens, err := m.db.GetNonExpiredTokensOfVersion(ctx, m.version)
	if err != nil {
		return fmt.Errorf("get active %s tokens from DB: %w", m.version, err)
	}

	activeTokenStrings := slices2.Map(activeTokens, func(t models.VmedisToken) string {
		return t.Token
	})

	m.activeTokensLock.Lock()
	m.activeTokens = activeTokenStrings
	m.activeTokensLock.Unlock()

	log.Printf("Finished reloading, got %d active %s tokens", len(activeTokens), m.version)

	return nil
}

func (m *Provider) Close() {
	m.closeOnce.Do(func() {
		close(m.closeCh)
	})
}

// NewProvider creates a Provider of the tokens of the given version.
// Having no such tokens yet is not an error: a token's version is only known
// once it is refreshed, and refreshing needs a client, which needs a provider.
func NewProvider(db *gorm.DB, version models.TokenVersion, reloadInterval time.Duration) (*Provider, error) {
	provider := &Provider{
		db:             NewDatabase(db),
		version:        version,
		reloadInterval: reloadInterval,
		closeCh:        make(chan struct{}),
	}

	if err := provider.ReloadTokens(context.Background()); err != nil {
		return nil, fmt.Errorf("initialize tokens: %w", err)
	}

	go provider.startReloader()

	return provider, nil
}
