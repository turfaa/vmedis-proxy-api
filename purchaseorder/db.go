package purchaseorder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// ErrNotFound is returned when a purchase order or its supplier does not
// exist.
var ErrNotFound = errors.New("not found")

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

// GetSupplier returns the supplier with the vmedis ID, if it is not
// soft-deleted.
func (d *Database) GetSupplier(ctx context.Context, vmedisID int) (models.Supplier, error) {
	var supplier models.Supplier
	if err := d.dbCtx(ctx).Where("vmedis_id = ?", vmedisID).Take(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Supplier{}, fmt.Errorf("supplier %d: %w", vmedisID, ErrNotFound)
		}
		return models.Supplier{}, fmt.Errorf("get supplier %d: %w", vmedisID, err)
	}

	return supplier, nil
}

// Create takes the next sequence of the purchase order's template and month,
// sets its Sequence and Number with numberFor, and inserts it with its items,
// all in one transaction. The sequence row is updated atomically, so
// concurrent creates never get the same number, and a failed insert does not
// use a number up.
func (d *Database) Create(ctx context.Context, po *models.PurchaseOrder, numberFor func(sequence int) string) error {
	return d.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		var sequence int
		if err := tx.Raw(
			`INSERT INTO purchase_order_sequences (template_code, year, month, last_sequence)
			VALUES (?, ?, ?, 1)
			ON CONFLICT (template_code, year, month)
			DO UPDATE SET last_sequence = purchase_order_sequences.last_sequence + 1
			RETURNING last_sequence`,
			po.TemplateCode, po.Year, po.Month,
		).Scan(&sequence).Error; err != nil {
			return fmt.Errorf("take next sequence: %w", err)
		}

		if sequence == 0 {
			return errors.New("take next sequence: no sequence returned")
		}

		po.Sequence = sequence
		po.Number = numberFor(sequence)

		if err := tx.Create(po).Error; err != nil {
			return fmt.Errorf("insert purchase order: %w", err)
		}

		return nil
	})
}

// GetDrugs returns the drugs with the vmedis codes, with their units ordered
// from the smallest, by code. Codes of unknown drugs are left out.
func (d *Database) GetDrugs(ctx context.Context, codes []string) (map[string]models.Drug, error) {
	drugs := make(map[string]models.Drug, len(codes))
	if len(codes) == 0 {
		return drugs, nil
	}

	var rows []models.Drug
	if err := d.dbCtx(ctx).
		Preload("Units", func(db *gorm.DB) *gorm.DB { return db.Order("unit_order") }).
		Where("vmedis_code IN ?", codes).
		Find(&rows).
		Error; err != nil {
		return nil, fmt.Errorf("get drugs: %w", err)
	}

	for _, drug := range rows {
		drugs[drug.VmedisCode] = drug
	}

	return drugs, nil
}

type cancellation struct {
	at           time.Time
	reason       string
	employeeID   int
	employeeName string
	email        string
}

// Cancel marks the purchase order as cancelled. The update only applies to
// a purchase order that is not cancelled yet, so of two concurrent cancels
// one wins and the other gets ErrAlreadyCancelled.
func (d *Database) Cancel(ctx context.Context, id uint, c cancellation) error {
	res := d.dbCtx(ctx).
		Model(&models.PurchaseOrder{}).
		Where("id = ? AND cancelled_at IS NULL", id).
		Updates(map[string]any{
			"cancelled_at":               c.at,
			"cancelled_reason":           c.reason,
			"cancelled_by_employee_id":   c.employeeID,
			"cancelled_by_employee_name": c.employeeName,
			"cancelled_by_email":         c.email,
		})
	if res.Error != nil {
		return fmt.Errorf("cancel purchase order %d: %w", id, res.Error)
	}

	if res.RowsAffected > 0 {
		return nil
	}

	var count int64
	if err := d.dbCtx(ctx).Model(&models.PurchaseOrder{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("check purchase order %d: %w", id, err)
	}

	if count == 0 {
		return fmt.Errorf("purchase order %d: %w", id, ErrNotFound)
	}

	return fmt.Errorf("purchase order %d: %w", id, ErrAlreadyCancelled)
}

// GetByIdempotencyKey returns the purchase order created with the key,
// without its items.
func (d *Database) GetByIdempotencyKey(ctx context.Context, key string) (models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := d.dbCtx(ctx).Where("idempotency_key = ?", key).Take(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PurchaseOrder{}, fmt.Errorf("purchase order with idempotency key %q: %w", key, ErrNotFound)
		}
		return models.PurchaseOrder{}, fmt.Errorf("get purchase order by idempotency key: %w", err)
	}

	return po, nil
}

// Get returns the purchase order with its items, ordered by position.
func (d *Database) Get(ctx context.Context, id uint) (models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := d.dbCtx(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Take(&po, id).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PurchaseOrder{}, fmt.Errorf("purchase order %d: %w", id, ErrNotFound)
		}
		return models.PurchaseOrder{}, fmt.Errorf("get purchase order %d: %w", id, err)
	}

	return po, nil
}

// List returns a page of the purchase orders matching the filter, newest
// first, the number of items of each, and how many purchase orders match in
// total.
func (d *Database) List(ctx context.Context, filter ListFilter) (pos []models.PurchaseOrder, itemCounts map[uint]int, total int64, err error) {
	if err := d.dbCtx(ctx).
		Model(&models.PurchaseOrder{}).
		Scopes(filterScope(filter)).
		Count(&total).
		Error; err != nil {
		return nil, nil, 0, fmt.Errorf("count purchase orders: %w", err)
	}

	if err := d.dbCtx(ctx).
		Scopes(filterScope(filter)).
		Order("created_at DESC").
		Order("id DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&pos).
		Error; err != nil {
		return nil, nil, 0, fmt.Errorf("list purchase orders: %w", err)
	}

	itemCounts, err = d.countItems(ctx, pos)
	if err != nil {
		return nil, nil, 0, err
	}

	return pos, itemCounts, total, nil
}

func filterScope(filter ListFilter) func(*gorm.DB) *gorm.DB {
	return func(query *gorm.DB) *gorm.DB {
		if q := strings.TrimSpace(filter.Query); q != "" {
			pattern := "%" + escapeLike(strings.ToLower(q)) + "%"
			query = query.Where(
				`(LOWER(number) LIKE ? ESCAPE '\' OR id IN (SELECT purchase_order_id FROM purchase_order_items WHERE LOWER(drug_name) LIKE ? ESCAPE '\'))`,
				pattern, pattern,
			)
		}

		if !filter.From.IsZero() {
			query = query.Where("created_at >= ?", filter.From)
		}

		if !filter.Until.IsZero() {
			query = query.Where("created_at <= ?", filter.Until)
		}

		if filter.SupplierVmedisID != 0 {
			query = query.Where("supplier_vmedis_id = ?", filter.SupplierVmedisID)
		}

		if filter.TemplateID != "" {
			query = query.Where("template_code = ?", filter.TemplateID)
		}

		if filter.CreatedByEmployeeID != 0 {
			query = query.Where("created_by_employee_id = ?", filter.CreatedByEmployeeID)
		}

		switch filter.Status {
		case StatusActive:
			query = query.Where("cancelled_at IS NULL")
		case StatusCancelled:
			query = query.Where("cancelled_at IS NOT NULL")
		}

		return query
	}
}

func (d *Database) countItems(ctx context.Context, pos []models.PurchaseOrder) (map[uint]int, error) {
	counts := make(map[uint]int, len(pos))
	if len(pos) == 0 {
		return counts, nil
	}

	ids := make([]uint, len(pos))
	for i, po := range pos {
		ids[i] = po.ID
	}

	var rows []struct {
		PurchaseOrderID uint
		Count           int
	}
	if err := d.dbCtx(ctx).
		Model(&models.PurchaseOrderItem{}).
		Select("purchase_order_id, COUNT(*) AS count").
		Where("purchase_order_id IN ?", ids).
		Group("purchase_order_id").
		Scan(&rows).
		Error; err != nil {
		return nil, fmt.Errorf("count purchase order items: %w", err)
	}

	for _, row := range rows {
		counts[row.PurchaseOrderID] = row.Count
	}

	return counts, nil
}

// ListCreators returns the employees who have created at least one purchase
// order, with the name of their latest one, ordered by name.
func (d *Database) ListCreators(ctx context.Context) ([]Employee, error) {
	var creators []Employee
	if err := d.dbCtx(ctx).
		Model(&models.PurchaseOrder{}).
		Select("created_by_employee_id AS id, created_by_employee_name AS name").
		Where("id IN (?)", d.dbCtx(ctx).
			Model(&models.PurchaseOrder{}).
			Select("MAX(id)").
			Group("created_by_employee_id"),
		).
		Order("created_by_employee_name").
		Scan(&creators).
		Error; err != nil {
		return nil, fmt.Errorf("list purchase order creators: %w", err)
	}

	return creators, nil
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// escapeLike escapes the LIKE wildcards in s, for use with ESCAPE '\'.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
