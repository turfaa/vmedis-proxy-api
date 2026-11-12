package procurement

import (
	"gorm.io/datatypes"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func vmedisProcurementToDBProcurement(p vmedisv2.Procurement) models.Procurement {
	return models.Procurement{
		InvoiceNumber:          p.InvoiceNumber,
		InvoiceDate:            datatypes.Date(p.Date),
		InputDate:              p.InputTime,
		Supplier:               p.Supplier,
		Warehouse:              p.Warehouse,
		PaymentType:            p.PaymentType,
		PaymentAccount:         p.PaymentAccount,
		Operator:               p.Operator,
		CashDiscountPercentage: p.CashDiscountPercentage,
		DiscountPercentage:     p.DiscountPercentage,
		DiscountAmount:         p.DiscountAmount,
		TaxPercentage:          p.TaxPercentage,
		TaxAmount:              p.TaxAmount,
		MiscellaneousCost:      p.MiscellaneousCost,
		Total:                  p.Total,
		ProcurementUnits: slices2.Map(p.ProcurementUnits, func(u vmedisv2.ProcurementUnit) models.ProcurementUnit {
			unit := vmedisProcurementUnitToDBProcurementUnit(u)
			unit.InvoiceNumber = p.InvoiceNumber
			return unit
		}),
	}
}

func vmedisProcurementUnitToDBProcurementUnit(u vmedisv2.ProcurementUnit) models.ProcurementUnit {
	return models.ProcurementUnit{
		IDInProcurement:         u.IDInProcurement,
		DrugCode:                u.DrugCode,
		DrugName:                u.DrugName,
		Amount:                  u.Amount,
		Unit:                    u.Unit,
		UnitBasePrice:           u.UnitBasePrice,
		DiscountPercentage:      u.DiscountPercentage,
		DiscountTwoPercentage:   u.DiscountTwoPercentage,
		DiscountThreePercentage: u.DiscountThreePercentage,
		TotalUnitPrice:          u.TotalUnitPrice,
		UnitTaxedPrice:          u.UnitTaxedPrice,
		ExpiryDate:              u.ExpiryDate,
		BatchNumber:             u.BatchNumber,
		Total:                   u.Total,
	}
}
