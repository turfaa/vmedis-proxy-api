package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"

	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func (d *Database) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	if err := d.db.Where(User{Email: email}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, fmt.Errorf("user with email %s not found", email)
		}

		return models.User{}, fmt.Errorf("get from db: %w", err)
	}

	return user, nil
}

func (d *Database) GetOrCreateUser(ctx context.Context, email string) (models.User, error) {
	now := time.Now()

	var user models.User
	if err := d.db.Where(models.User{Email: email}).
		Attrs(models.User{
			Role:      "guest",
			CreatedAt: now,
			UpdatedAt: now,
		}).
		FirstOrCreate(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("get or create from db: %w", err)
	}

	return user, nil
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}
