package sale

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *Database
}

func (s *Service) GetAggregatedSalesBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]AggregatedSale, error) {
	return s.db.GetAggregatedSalesBetweenTime(ctx, from, to)
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: NewDatabase(db),
	}
}
