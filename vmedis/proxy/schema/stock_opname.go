package schema

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// StockOpnamesResponse represents the stock opnames API response.
type StockOpnamesResponse struct {
	StockOpnames []StockOpname `json:"stockOpnames"`
}

// StockOpname represents the stock opname.
type StockOpname struct {
	VmedisID            string  `json:"vmedisId"`
	Date                string  `json:"date"`
	DrugCode            string  `json:"drugCode"`
	DrugName            string  `json:"drugName"`
	BatchCode           string  `json:"batchCode"`
	Unit                string  `json:"unit"`
	InitialQuantity     float64 `json:"initialQuantity"`
	RealQuantity        float64 `json:"realQuantity"`
	QuantityDifference  float64 `json:"quantityDifference"`
	HPPDifference       float64 `json:"hppDifference"`
	SalePriceDifference float64 `json:"salePriceDifference"`
	Notes               string  `json:"notes"`
}

// FromModelsStockOpname converts StockOpname from models.StockOpname to proxy schema.
func FromModelsStockOpname(stockOpname models.StockOpname) StockOpname {
	return StockOpname{
		VmedisID:            stockOpname.VmedisID,
		Date:                time.Time(stockOpname.Date).Format("2006-01-02"),
		DrugCode:            stockOpname.DrugCode,
		DrugName:            stockOpname.DrugName,
		BatchCode:           stockOpname.BatchCode,
		Unit:                stockOpname.Unit,
		InitialQuantity:     stockOpname.InitialQuantity,
		RealQuantity:        stockOpname.RealQuantity,
		QuantityDifference:  stockOpname.QuantityDifference,
		HPPDifference:       stockOpname.HPPDifference,
		SalePriceDifference: stockOpname.SalePriceDifference,
		Notes:               stockOpname.Notes,
	}
}
