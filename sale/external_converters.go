package sale

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func VmedisSaleToDBSale(sale vmedis.Sale) models.Sale {
	return models.Sale{
		VmedisID:      sale.ID,
		SoldAt:        sale.Date.Time,
		InvoiceNumber: sale.InvoiceNumber,
		PatientName:   sale.PatientName,
		Doctor:        sale.Doctor,
		Payment:       sale.Payment,
		Total:         sale.Total,
		SaleUnits: slices2.Map(sale.SaleUnits, func(saleUnit vmedis.SaleUnit) models.SaleUnit {
			su := VmedisSaleUnitToDBSaleUnit(saleUnit)
			su.InvoiceNumber = sale.InvoiceNumber
			return su
		}),
	}
}

func VmedisSaleUnitToDBSaleUnit(saleUnit vmedis.SaleUnit) models.SaleUnit {
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
