package token

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type Database struct {
	db *gorm.DB
}

func (d *Database) Transaction(ctx context.Context, f func(d *Database) error) error {
	tx := d.withContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}

	if err := f(NewDatabase(tx)); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (d *Database) GetNonExpiredTokens(ctx context.Context) ([]models.VmedisToken, error) {
	var tokens []models.VmedisToken
	if err := d.withContext(ctx).Where("state != 'EXPIRED'").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("get non expired tokens from DB: %w", err)
	}

	return tokens, nil
}

func (d *Database) GetActiveTokens(ctx context.Context) ([]models.VmedisToken, error) {
	var tokens []models.VmedisToken
	if err := d.withContext(ctx).Where("state = 'ACTIVE'").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("get active tokens from DB: %w", err)
	}

	return tokens, nil
}

func (d *Database) UpsertTokensState(ctx context.Context, tokens []models.VmedisToken) error {
	if len(tokens) == 0 {
		return nil
	}

	// Clear auto-assignable fields.
	for i := range tokens {
		tokens[i].ID = 0
		tokens[i].CreatedAt = time.Time{}
		tokens[i].UpdatedAt = time.Time{}
	}

	if err := d.withContext(ctx).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "token"}},
				DoUpdates: clause.AssignmentColumns([]string{"updated_at", "state"}),
			},
		).
		Create(&tokens).
		Error; err != nil {
		return fmt.Errorf("upsert tokens state: %w", err)
	}

	return nil
}

func (d *Database) withContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}
