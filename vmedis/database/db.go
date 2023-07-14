package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// SqliteDB returns the sqlite database.
func SqliteDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

// AutoMigrate auto migrates available models.
func AutoMigrate(db *gorm.DB) error {
	availableModels := []interface{}{
		models.SaleStatistics{},
	}

	for _, model := range availableModels {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("auto migrate %T: %w", model, err)
		}
	}

	return nil
}
