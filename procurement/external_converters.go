package procurement

import (
	"gorm.io/datatypes"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func vmedisProcurementToDBProcurement(p vmedis.Procurement) models.Procurement {
	return models.Procurement{
		InvoiceNumber:          p.InvoiceNumber,
		InvoiceDate:            datatypes.Date(p.Date.Time),
		Supplier:               p.Supplier,
		Warehouse:              p.Warehouse,
		PaymentType:            p.PaymentType,
		Operator:               p.Operator,
		CashDiscountPercentage: p.CashDiscountPercentage.Value,
		DiscountPercentage:     p.DiscountPercentage.Value,
		DiscountAmount:         p.DiscountAmount,
		TaxPercentage:          p.TaxPercentage.Value,
		TaxAmount:              p.TaxAmount,
		MiscellaneousCost:      p.MiscellaneousCost,
		Total:                  p.Total,
		ProcurementUnits: slices2.Map(p.ProcurementUnits, func(u vmedis.ProcurementUnit) models.ProcurementUnit {
			unit := vmedisProcurementUnitToDBProcurementUnit(u)
			unit.InvoiceNumber = p.InvoiceNumber
			return unit
		}),
	}
}

func vmedisProcurementUnitToDBProcurementUnit(u vmedis.ProcurementUnit) models.ProcurementUnit {
	return models.ProcurementUnit{
		IDInProcurement:         u.IDInProcurement,
		DrugCode:                u.DrugCode,
		DrugName:                u.DrugName,
		Amount:                  u.Amount,
		Unit:                    u.Unit,
		UnitBasePrice:           u.UnitBasePrice,
		DiscountPercentage:      u.DiscountPercentage.Value,
		DiscountTwoPercentage:   u.DiscountTwoPercentage.Value,
		DiscountThreePercentage: u.DiscountThreePercentage.Value,
		TotalUnitPrice:          u.TotalUnitPrice,
		UnitTaxedPrice:          u.UnitTaxedPrice,
		ExpiryDate:              u.ExpiryDate.Time,
		BatchNumber:             u.BatchNumber,
		Total:                   u.Total,
	}
}
