package drug

import (
	"context"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// Database provides drug-related database operations.
type Database struct {
	db *gorm.DB
}

// GetDrugsUpdatedAfter returns the drugs updated after the given time.
func (d *Database) GetDrugsUpdatedAfter(ctx context.Context, minimumUpdatedTime time.Time) ([]models.Drug, error) {
	return d.getDrugsUpdatedAfterWithAdditionalQuery(ctx, minimumUpdatedTime, nil)
}

// GetDrugsByVmedisCodesUpdatedAfter returns drugs with the given vmedis codes
// updated after the given time.
func (d *Database) GetDrugsByVmedisCodesUpdatedAfter(
	ctx context.Context,
	vmedisCodes []string,
	minimumUpdatedTime time.Time,
) ([]models.Drug, error) {
	return d.getDrugsUpdatedAfterWithAdditionalQuery(
		ctx,
		minimumUpdatedTime,
		func(db *gorm.DB) *gorm.DB {
			return db.Where("vmedis_code IN ?", vmedisCodes)
		},
	)
}

// GetDrugsByExcludedVmedisCodesUpdatedAfter returns drugs which code not in the given vmedis codes
// updated after the given time.
func (d *Database) GetDrugsByExcludedVmedisCodesUpdatedAfter(
	ctx context.Context,
	excludedVmedisCodes []string,
	minimumUpdatedTime time.Time,
) ([]models.Drug, error) {
	return d.getDrugsUpdatedAfterWithAdditionalQuery(
		ctx,
		minimumUpdatedTime,
		func(db *gorm.DB) *gorm.DB {
			if len(excludedVmedisCodes) == 0 {
				return db
			}

			return db.Where("vmedis_code NOT IN ?", excludedVmedisCodes)
		},
	)
}

func (d *Database) getDrugsUpdatedAfterWithAdditionalQuery(
	ctx context.Context,
	minimumUpdatedTime time.Time,
	additionalQuery func(db *gorm.DB) *gorm.DB,
) ([]models.Drug, error) {
	if additionalQuery == nil {
		additionalQuery = func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	query := additionalQuery(
		d.dbCtx(ctx).
			Preload("Units").
			Preload("Stocks").
			Where("updated_at >= ?", minimumUpdatedTime).
			Order("name"),
	)

	var drugs []models.Drug
	if err := query.Find(&drugs).Error; err != nil {
		return nil, fmt.Errorf("get drugs updated after %s: %w", minimumUpdatedTime, err)
	}

	return drugs, nil
}

// GetDrugCodesAlreadyStockOpnamedBetweenTimes returns the vmedis code of the drugs already stock opnamed between the given times.
func (d *Database) GetDrugCodesAlreadyStockOpnamedBetweenTimes(
	ctx context.Context,
	startTime time.Time,
	endTime time.Time,
) ([]string, error) {
	query := d.dbCtx(ctx).
		Model(&models.StockOpname{}).
		Where("date BETWEEN ? AND ?", startTime, endTime)

	var drugCodes []string
	if err := query.Pluck("drug_code", &drugCodes).Error; err != nil {
		return nil, fmt.Errorf("get drug codes already stock opnamed between %s and %s: %w", startTime, endTime, err)
	}

	return drugCodes, nil
}

// GetDrugSaleStatisticsBetweenTimes returns the drug sale statistics between the given times.
func (d *Database) GetDrugSaleStatisticsBetweenTimes(
	ctx context.Context,
	startTime time.Time,
	endTime time.Time,
) ([]SaleStatistics, error) {
	query := d.dbCtx(ctx).
		Model(&models.SaleUnit{}).
		Select("drug_code, COUNT(*) AS number_of_sales, SUM(total) AS total_amount").
		Where("invoice_number IN (SELECT invoice_number FROM sales WHERE sold_at BETWEEN ? AND ?)", startTime, endTime).
		Group("drug_code")

	var stats []SaleStatistics
	if err := query.Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("get drug sale statistics between %s and %s: %w", startTime, endTime, err)
	}

	return stats, nil
}

// GetDrugUnitsByDrugVmedisCodes returns drug units by drug vmedis codes.
// The drug units are sorted from the smallest to the largest.
func (d *Database) GetDrugUnitsByDrugVmedisCodes(ctx context.Context, drugVmedisCodes []string) (map[string][]Unit, error) {
	var units []models.DrugUnit
	if err := d.dbCtx(ctx).Where("drug_vmedis_code IN ?", drugVmedisCodes).Find(&units).Error; err != nil {
		return nil, fmt.Errorf("get drug units by drug vmedis codes %v: %w", drugVmedisCodes, err)
	}

	unitsByDrugVmedisCode := make(map[string][]Unit, len(drugVmedisCodes))
	for _, unit := range units {
		unitsByDrugVmedisCode[unit.DrugVmedisCode] = append(unitsByDrugVmedisCode[unit.DrugVmedisCode], FromDBDrugUnit(unit))
	}

	for drugVmedisCode, drugUnits := range unitsByDrugVmedisCode {
		sorted := make([]Unit, 0, len(drugUnits))

		last := ""
		for {
			found := false
			for _, u := range drugUnits {
				if u.ParentUnit == last {
					sorted = append(sorted, u)
					last = u.Unit
					found = true
					break
				}
			}

			if !found {
				break
			}
		}

		unitsByDrugVmedisCode[drugVmedisCode] = sorted
	}

	return unitsByDrugVmedisCode, nil
}

// UpsertVmedisDrug upserts the given drug.
func (d *Database) UpsertVmedisDrug(ctx context.Context, drug vmedis.Drug, keyColumn string, updateColumns []string) error {
	return d.UpsertVmedisDrugs(ctx, []vmedis.Drug{drug}, keyColumn, updateColumns)
}

// UpsertVmedisDrugs upserts the given drugs.
func (d *Database) UpsertVmedisDrugs(ctx context.Context, drugs []vmedis.Drug, keyColumn string, updateColumns []string) error {
	if len(drugs) == 0 {
		return nil
	}

	dbDrugs := slices2.Map(drugs, func(drug vmedis.Drug) models.Drug {
		return models.Drug{
			VmedisID:     drug.VmedisID,
			VmedisCode:   drug.VmedisCode,
			Name:         drug.Name,
			Manufacturer: drug.Manufacturer,
			MinimumStock: models.Stock{
				Unit:     drug.MinimumStock.Unit,
				Quantity: drug.MinimumStock.Quantity,
			},
		}
	})

	if !slices.Contains(updateColumns, "updated_at") {
		updateColumns = append(updateColumns, "updated_at")
	}

	ops := d.dbCtx(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: keyColumn}},
			DoUpdates: clause.AssignmentColumns(updateColumns),
		}).
		Create(&dbDrugs)

	if err := ops.Error; err != nil {
		return fmt.Errorf("upsert vmedis drugs: %w", err)
	}

	return nil
}

// UpsertVmedisDrugUnits upserts the given drug units.
func (d *Database) UpsertVmedisDrugUnits(ctx context.Context, drugVmedisCode string, units []vmedis.Unit) error {
	if len(units) == 0 {
		return nil
	}

	dbUnits := slices2.Map(units, func(unit vmedis.Unit) models.DrugUnit {
		return models.DrugUnit{
			DrugVmedisCode:         drugVmedisCode,
			Unit:                   unit.Unit,
			ParentUnit:             unit.ParentUnit,
			ConversionToParentUnit: unit.ConversionToParentUnit,
			PriceOne:               unit.PriceOne,
			PriceTwo:               unit.PriceTwo,
			PriceThree:             unit.PriceThree,
			UnitOrder:              unit.UnitOrder,
		}
	})

	ops := d.dbCtx(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "drug_vmedis_code"}, {Name: "unit"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at",
				"parent_unit",
				"conversion_to_parent_unit",
				"unit_order",
				"price_one",
				"price_two",
				"price_three",
			}),
		}).
		Create(&dbUnits)

	if err := ops.Error; err != nil {
		return fmt.Errorf("upsert vmedis drug units: %w", err)
	}

	return nil
}

// UpsertVmedisDrugStocks upserts the given drug stocks.
func (d *Database) UpsertVmedisDrugStocks(ctx context.Context, drugVmedisCode string, stocks []vmedis.Stock) error {
	dbStocks := slices2.Map(stocks, func(stock vmedis.Stock) models.DrugStock {
		return models.DrugStock{
			DrugVmedisCode: drugVmedisCode,
			Stock: models.Stock{
				Unit:     stock.Unit,
				Quantity: stock.Quantity,
			},
		}
	})

	if err := d.dbCtx(ctx).
		Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(models.DrugStock{}, "drug_vmedis_code = ?", drugVmedisCode).Error; err != nil {
				return fmt.Errorf("delete drug stocks of '%s': %w", drugVmedisCode, err)
			}

			if len(dbStocks) == 0 {
				return nil
			}

			if err := tx.Create(&dbStocks).Error; err != nil {
				return fmt.Errorf("create drug stocks: %w", err)
			}

			return nil
		}); err != nil {
		return fmt.Errorf("upsert vmedis drug stocks transaction: %w", err)
	}

	return nil
}

func (d *Database) dbCtx(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// NewDatabase creates a new drug database.
func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}
