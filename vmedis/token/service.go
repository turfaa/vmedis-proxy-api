package token

import (
	"context"
	"fmt"
	"strings"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"gorm.io/gorm"
)

type Service struct {
	db *Database
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: NewDatabase(db)}
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
