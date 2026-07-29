package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// UndeleteAndUpdateColumns builds the assignments of an upsert's DO UPDATE
// clause for a soft-deletable model: it refreshes the given columns from the
// row being inserted, and clears deleted_at.
//
// The unique constraints of the soft-deletable models span their soft-deleted
// rows, so an upsert of a record that was soft-deleted here conflicts with the
// hidden row rather than inserting a second one. Clearing deleted_at is what
// makes that upsert bring the record back to life instead of quietly
// refreshing a row nothing can see.
func UndeleteAndUpdateColumns(columns []string) clause.Set {
	return append(
		clause.AssignmentColumns(columns),
		clause.Assignment{Column: clause.Column{Name: "deleted_at"}, Value: gorm.Expr("NULL")},
	)
}

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
		models.RejectedDrug{},
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
