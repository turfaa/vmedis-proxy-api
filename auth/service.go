package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Service struct {
	db *Database
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return User{
		Email: user.Email,
		Role:  Role(user.Role),
	}, nil
}

func (s *Service) GetOrCreateUser(ctx context.Context, email string) (User, error) {
	user, err := s.db.GetOrCreateUser(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("get or create user: %w", err)
	}

	return User{
		Email: user.Email,
		Role:  Role(user.Role),
	}, nil
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: NewDatabase(db)}
}
