package supplier

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

// UpsertVmedisSuppliers inserts the suppliers, or updates them by their
// vmedis IDs, reviving the ones that were soft-deleted.
func (d *Database) UpsertVmedisSuppliers(ctx context.Context, suppliers []vmedisv2.Supplier) error {
	if len(suppliers) == 0 {
		return nil
	}

	dbSuppliers := slices2.Map(suppliers, vmedisSupplierToDBSupplier)

	if err := d.dbCtx(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "vmedis_id"}},
			DoUpdates: database.UndeleteAndUpdateColumns([]string{
				"updated_at",
				"code",
				"name",
				"address",
				"city",
				"phone",
				"mobile_phone",
				"bank_account_number",
				"tax_number",
				"active",
				"lead_time_days",
				"rating",
			}),
		}).
		CreateInBatches(&dbSuppliers, 500).
		Error; err != nil {
		return fmt.Errorf("upsert suppliers: %w", err)
	}

	return nil
}

// SoftDeleteSuppliersNotIn soft-deletes the suppliers whose vmedis IDs are not
// in vmedisIDs, and returns how many were soft-deleted.
func (d *Database) SoftDeleteSuppliersNotIn(ctx context.Context, vmedisIDs []int) (int64, error) {
	res := d.dbCtx(ctx).Where("vmedis_id NOT IN ?", vmedisIDs).Delete(&models.Supplier{})
	if res.Error != nil {
		return 0, fmt.Errorf("soft-delete suppliers not in vmedis: %w", res.Error)
	}

	return res.RowsAffected, nil
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}
