package schema

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// SalesResponse represents the sales API response.
type SalesResponse struct {
	Sales []Sale `json:"sales"`
}

// SoldDrugsResponse represents the sold drugs API response.
type SoldDrugsResponse struct {
	Date  string `json:"date"`
	Drugs []Drug `json:"drugs"`
}

// Sale represents a sale.
type Sale struct {
	VmedisID      int        `json:"vmedisId"`
	SoldAt        time.Time  `json:"soldAt"`
	InvoiceNumber string     `json:"invoiceNumber"`
	PatientName   string     `json:"patientName,omitempty"`
	Doctor        string     `json:"doctor,omitempty"`
	Payment       string     `json:"payment"`
	Total         float64    `json:"total"`
	SaleUnits     []SaleUnit `json:"saleUnits"`
}

// FromModelsSale converts Sale from models.Sale to proxy schema.
func FromModelsSale(sale models.Sale) Sale {
	sus := make([]SaleUnit, len(sale.SaleUnits))
	for i, su := range sale.SaleUnits {
		sus[i] = FromModelsSaleUnit(su)
	}

	return Sale{
		VmedisID:      sale.VmedisID,
		SoldAt:        sale.SoldAt,
		InvoiceNumber: sale.InvoiceNumber,
		PatientName:   sale.PatientName,
		Doctor:        sale.Doctor,
		Payment:       sale.Payment,
		Total:         sale.Total,
		SaleUnits:     sus,
	}
}

// SaleUnit represents one unit of a drug in a sale.
type SaleUnit struct {
	IDInSale      int     `json:"idInSale"`
	DrugCode      string  `json:"drugCode"`
	DrugName      string  `json:"drugName"`
	Batch         string  `json:"batch"`
	Amount        float64 `json:"amount"`
	Unit          string  `json:"unit"`
	UnitPrice     float64 `json:"unitPrice"`
	PriceCategory string  `json:"priceCategory"`
	Discount      float64 `json:"discount,omitempty"`
	Tuslah        float64 `json:"tuslah,omitempty"`
	Embalase      float64 `json:"embalase,omitempty"`
	Total         float64 `json:"total"`
}

// FromModelsSaleUnit converts SaleUnit from models.SaleUnit to proxy schema.
func FromModelsSaleUnit(saleUnit models.SaleUnit) SaleUnit {
	return SaleUnit{
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
