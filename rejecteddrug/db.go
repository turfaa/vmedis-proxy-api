package rejecteddrug

import (
	"context"
	"fmt"
	"strings"

	"github.com/turfaa/vmedis-proxy-api/database/models"

	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

func (d *Database) GetRejectedDrugs(ctx context.Context, filters ListFilters) ([]models.RejectedDrug, error) {
	var rejectedDrugs []models.RejectedDrug

	if err := applyFilters(d.dbCtx(ctx), filters).
		Order("created_at DESC").
		Find(&rejectedDrugs).
		Error; err != nil {
		return nil, fmt.Errorf("get rejected drugs from db: %w", err)
	}

	return rejectedDrugs, nil
}

func (d *Database) GetRejectedDrugByID(ctx context.Context, id uint) (models.RejectedDrug, error) {
	var rejectedDrug models.RejectedDrug

	if err := d.dbCtx(ctx).First(&rejectedDrug, id).Error; err != nil {
		return models.RejectedDrug{}, fmt.Errorf("get rejected drug %d from db: %w", id, err)
	}

	return rejectedDrug, nil
}

func (d *Database) CreateRejectedDrug(ctx context.Context, rejectedDrug models.RejectedDrug) (models.RejectedDrug, error) {
	if err := d.dbCtx(ctx).Create(&rejectedDrug).Error; err != nil {
		return models.RejectedDrug{}, fmt.Errorf("create rejected drug: %w", err)
	}

	return rejectedDrug, nil
}

func (d *Database) SaveRejectedDrug(ctx context.Context, rejectedDrug models.RejectedDrug) (models.RejectedDrug, error) {
	if err := d.dbCtx(ctx).Save(&rejectedDrug).Error; err != nil {
		return models.RejectedDrug{}, fmt.Errorf("save rejected drug %d: %w", rejectedDrug.ID, err)
	}

	return rejectedDrug, nil
}

func (d *Database) DeleteRejectedDrug(ctx context.Context, id uint) error {
	if err := d.dbCtx(ctx).Delete(&models.RejectedDrug{}, id).Error; err != nil {
		return fmt.Errorf("delete rejected drug %d: %w", id, err)
	}

	return nil
}

func applyFilters(query *gorm.DB, filters ListFilters) *gorm.DB {
	if filters.Query != "" {
		pattern := likePattern(filters.Query)
		query = query.Where(
			"LOWER(drug_name) LIKE ? OR LOWER(reason) LIKE ? OR LOWER(resolution_notes) LIKE ?",
			pattern, pattern, pattern,
		)
	}

	if filters.DrugName != "" {
		query = query.Where("LOWER(drug_name) LIKE ?", likePattern(filters.DrugName))
	}

	if filters.Reason != "" {
		query = query.Where("LOWER(reason) LIKE ?", likePattern(filters.Reason))
	}

	if filters.ResolutionNotes != "" {
		query = query.Where("LOWER(resolution_notes) LIKE ?", likePattern(filters.ResolutionNotes))
	}

	if len(filters.Resolutions) > 0 {
		query = query.Where("resolution IN ?", filters.Resolutions)
	}

	if filters.CreatedBy != "" {
		query = query.Where("created_by = ?", filters.CreatedBy)
	}

	if filters.ResolvedBy != "" {
		query = query.Where("resolved_by = ?", filters.ResolvedBy)
	}

	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filters.CreatedFrom)
	}

	if filters.CreatedUntil != nil {
		query = query.Where("created_at <= ?", *filters.CreatedUntil)
	}

	if filters.ResolvedFrom != nil {
		query = query.Where("resolved_at >= ?", *filters.ResolvedFrom)
	}

	if filters.ResolvedUntil != nil {
		query = query.Where("resolved_at <= ?", *filters.ResolvedUntil)
	}

	return query
}

func likePattern(value string) string {
	return "%" + strings.ToLower(value) + "%"
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}
