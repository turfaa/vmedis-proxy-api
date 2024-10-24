package shift

import (
	"context"
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

func (d *Database) GetShiftByCode(ctx context.Context, code string) (models.Shift, error) {
	var shift models.Shift

	if err := d.dbCtx(ctx).Where("code = ?", code).Order("started_at DESC").First(&shift).Error; err != nil {
		return models.Shift{}, fmt.Errorf("failed to get shift by code %s: %w", code, err)
	}

	return shift, nil
}

func (d *Database) GetShiftByVmedisID(ctx context.Context, vmedisID int) (models.Shift, error) {
	var shift models.Shift

	if err := d.dbCtx(ctx).Where("vmedis_id = ?", vmedisID).First(&shift).Error; err != nil {
		return models.Shift{}, fmt.Errorf("failed to get shift by vmedis id %d: %w", vmedisID, err)
	}

	return shift, nil
}

func (d *Database) GetShiftsBetween(ctx context.Context, from time.Time, to time.Time) ([]models.Shift, error) {
	var shifts []models.Shift

	if err := d.dbCtx(ctx).Where("started_at >= ? AND ended_at <= ?", from, to).Find(&shifts).Error; err != nil {
		return nil, fmt.Errorf("failed to get shifts between %s and %s from db: %w", from, to, err)
	}

	return shifts, nil
}

func (d *Database) UpsertVmedisShifts(ctx context.Context, shifts []vmedis.Shift) error {
	if len(shifts) == 0 {
		return nil
	}

	dbShifts := slices2.Map(shifts, vmedisShiftToDBShift)

	if err := d.dbCtx(ctx).
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "vmedis_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"updated_at",
					"code",
					"cashier",
					"started_at",
					"ended_at",
					"initial_cash",
					"expected_final_cash",
					"actual_final_cash",
					"final_cash_difference",
					"supervisor",
					"notes",
				}),
			},
		).
		Create(&dbShifts).
		Error; err != nil {
		return fmt.Errorf("failed to upsert shifts: %w", err)
	}

	return nil
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}
