package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
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

// PostgresDB returns the postgres database.
func PostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}

	if err := db.Exec(`
	DO $$ BEGIN
		CREATE TYPE token_state AS ENUM ('UNCHECKED', 'ACTIVE', 'EXPIRED');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`).
		Error; err != nil {
		return nil, fmt.Errorf("create token_state enum: %w", err)
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
		models.Drug{},
		models.DrugUnit{},
		models.DrugStock{},
		models.Sale{},
		models.SaleUnit{},
		models.StockOpname{},
		models.User{},
		models.InvoiceCalculator{},
		models.InvoiceComponent{},
		models.Procurement{},
		models.ProcurementUnit{},
		models.VmedisToken{},
		models.Shift{},
	}

	for _, model := range availableModels {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("auto migrate %T: %w", model, err)
		}
	}

	if err := PrepopulateInvoiceCalculators(db); err != nil {
		return fmt.Errorf("prepopulate invoice calculators: %w", err)
	}

	return nil
}
