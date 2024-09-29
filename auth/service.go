package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	cache *Cache
	db    *Database
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := s.cache.GetUser(ctx, email)
	if err == nil {
		return user, nil
	}

	userDB, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	user = User{
		Email: userDB.Email,
		Role:  Role(userDB.Role),
	}

	go func() {
		if err := s.cache.SetUser(context.Background(), user, time.Minute); err != nil {
			log.Printf("failed to set user to cache: %s", err)
		}
	}()

	return user, nil
}

func (s *Service) GetOrCreateUser(ctx context.Context, email string) (User, error) {
	user, err := s.cache.GetUser(ctx, email)
	if err == nil {
		return user, nil
	}

	userDB, err := s.db.GetOrCreateUser(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("get or create user: %w", err)
	}

	user = User{
		Email: userDB.Email,
		Role:  Role(userDB.Role),
	}

	go func() {
		if err := s.cache.SetUser(context.Background(), user, time.Minute); err != nil {
			log.Printf("failed to set user to cache: %s", err)
		}
	}()

	return user, nil
}

func NewService(redisClient *redis.Client, db *gorm.DB) *Service {
	return &Service{
		cache: NewCache(redisClient),
		db:    NewDatabase(db),
	}
}
