package token

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
)

type Service struct {
	db        *Database
	redisDB   *RedisDatabase
	refresher *Refresher
}

func NewService(db *gorm.DB, redisClient redis.UniversalClient, refresher *Refresher) *Service {
	return &Service{
		db:        NewDatabase(db),
		redisDB:   NewRedisDatabase(redisClient),
		refresher: refresher,
	}
}

func (s *Service) GetTokens(ctx context.Context) ([]models.VmedisToken, error) {
	tokens, err := s.db.GetAllTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all tokens from DB: %w", err)
	}

	sanitizedTokens := slices2.Map(tokens, func(t models.VmedisToken) models.VmedisToken {
		t.Token = s.censorToken(t.Token)
		return t
	})

	return sanitizedTokens, nil
}

func (*Service) censorToken(token string) string {
	length := len(token)
	halfLength := length / 2
	return token[:halfLength] + strings.Repeat("*", length-halfLength)
}

func (s *Service) InsertToken(ctx context.Context, token string) error {
	return s.db.InsertToken(ctx, token)
}

func (s *Service) DeleteToken(ctx context.Context, id uint) error {
	return s.db.DeleteToken(ctx, id)
}

// DeleteExpiredTokens deletes all expired tokens and returns the number deleted.
func (s *Service) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	return s.db.DeleteExpiredTokens(ctx)
}

// RefreshTokens refreshes the state of every non-expired token against Vmedis.
// It uses a distributed lock so only one refresh runs at a time; the lock is
// held for at most one minute. If a refresh is already in progress, it returns
// without doing anything.
func (s *Service) RefreshTokens(ctx context.Context) error {
	token, acquired, err := s.redisDB.AcquireRefreshLock(ctx)
	if err != nil {
		return fmt.Errorf("acquire token refresh lock: %w", err)
	}

	if !acquired {
		log.Println("Tokens are already being refreshed by another process, skipping")
		return nil
	}

	defer func() {
		if err := s.redisDB.ReleaseRefreshLock(context.WithoutCancel(ctx), token); err != nil {
			log.Printf("Failed to release token refresh lock: %s", err)
		}
	}()

	if err := s.refresher.RefreshTokens(ctx); err != nil {
		return fmt.Errorf("refresh tokens: %w", err)
	}

	return nil
}

// GetRefreshStatus reports whether a token refresh is currently in progress.
func (s *Service) GetRefreshStatus(ctx context.Context) (RefreshStatus, error) {
	locked, err := s.redisDB.IsRefreshLocked(ctx)
	if err != nil {
		return "", fmt.Errorf("get token refresh status: %w", err)
	}

	if locked {
		return RefreshStatusRefreshing, nil
	}

	return RefreshStatusIdle, nil
}
