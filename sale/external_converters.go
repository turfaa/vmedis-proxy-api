package sale

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

func VmedisSaleToDBSale(sale vmedisv1.Sale) models.Sale {
	return models.Sale{
		VmedisID:      sale.ID,
		SoldAt:        sale.Date.Time,
		Cashier:       sale.Cashier,
		InvoiceNumber: sale.InvoiceNumber,
		PatientName:   sale.PatientName,
		Doctor:        sale.Doctor,
		Salesman:      sale.Salesman,
		Payment:       sale.Payment,
		Total:         sale.Total,
		SaleUnits: slices2.Map(sale.SaleUnits, func(saleUnit vmedisv1.SaleUnit) models.SaleUnit {
			su := VmedisSaleUnitToDBSaleUnit(saleUnit)
			su.InvoiceNumber = sale.InvoiceNumber
			return su
		}),
	}
}

func VmedisSaleUnitToDBSaleUnit(saleUnit vmedisv1.SaleUnit) models.SaleUnit {
	return models.SaleUnit{
		IDInSale:      saleUnit.IDInSale,
		DrugCode:      saleUnit.DrugCode,
		DrugName:      saleUnit.DrugName,
		Batch:         saleUnit.Batch,
		Amount:        saleUnit.Amount,
		Unit:          saleUnit.Unit,
		UnitPrice:     saleUnit.UnitPrice,
		PriceCategory: saleUnit.PriceCategory,
		Discount:      saleUnit.Discount,
		Tuslah:        saleUnit.Tuslah,
		Embalase:      saleUnit.Embalase,
		Total:         saleUnit.Total,
	}
}
