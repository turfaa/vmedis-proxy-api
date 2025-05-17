package sale

import (
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/money"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type SalesResponse struct {
	Sales []Sale `json:"sales"`
}

type Sale struct {
	VmedisID      int       `json:"vmedisId"`
	SoldAt        time.Time `json:"soldAt"`
	InvoiceNumber string    `json:"invoiceNumber"`
	PatientName   string    `json:"patientName,omitempty"`
	Doctor        string    `json:"doctor,omitempty"`
	Payment       string    `json:"payment"`
	Total         float64   `json:"total"`
	SaleUnits     []Unit    `json:"saleUnits"`
}

func FromDBSale(sale models.Sale) Sale {
	sus := make([]Unit, len(sale.SaleUnits))
	for i, su := range sale.SaleUnits {
		sus[i] = FromDBSaleUnit(su)
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

type Unit struct {
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

func FromDBSaleUnit(saleUnit models.SaleUnit) Unit {
	return Unit{
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

type AggregatedSale struct {
	DrugName string  `json:"drugName"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type SoldDrugsResponse struct {
	Drugs []SoldDrug `json:"drugs"`
}

// SoldDrug represents a sold drug.
type SoldDrug struct {
	Drug        drug.Drug `json:"drug"`
	Occurrences int       `json:"occurrences"`
	TotalAmount float64   `json:"totalAmount"`
}

type StatisticsResponse struct {
	History      []Statistics `json:"history"`
	DailyHistory []Statistics `json:"dailyHistory"`
}

type Statistics struct {
	PulledAt      time.Time `json:"pulledAt"`
	TotalSales    float64   `json:"totalSales"`
	NumberOfSales int       `json:"numberOfSales"`
}

func (s Statistics) ToDBSaleStatistics() models.SaleStatistics {
	return models.SaleStatistics{
		PulledAt:      s.PulledAt,
		TotalSales:    s.TotalSales,
		NumberOfSales: s.NumberOfSales,
	}
}

func FromVmedisSalesStatistics(pulledAt time.Time, salesStatistics vmedis.SalesStatistics) (Statistics, error) {
	totalSales, err := salesStatistics.TotalSalesFloat64()
	if err != nil {
		return Statistics{}, fmt.Errorf("total sales float64: %w", err)
	}

	return Statistics{
		PulledAt:      pulledAt,
		TotalSales:    totalSales,
		NumberOfSales: salesStatistics.NumberOfSales,
	}, nil
}

func FromDBSaleStatistics(saleStatistics models.SaleStatistics) Statistics {
	return Statistics{
		PulledAt:      saleStatistics.PulledAt,
		TotalSales:    saleStatistics.TotalSales,
		NumberOfSales: saleStatistics.NumberOfSales,
	}
}

// StatisticsSensorsResponse is the response for the special API
// for creating Home Assistant sensors based on sale statistics.
type StatisticsSensorsResponse struct {
	Today     StatisticsSensorResponse `json:"today"`
	Yesterday StatisticsSensorResponse `json:"yesterday"`
}

type StatisticsSensorResponse struct {
	DateString string `json:"dateString"`
	TotalSales string `json:"totalSales"`
}

// StatisticsSensors represents all Home Assistant sensors for sale statistics.
type StatisticsSensors struct {
	Today     StatisticsSensor
	Yesterday StatisticsSensor
}

func (s StatisticsSensors) ToStatisticsSensorsResponse() StatisticsSensorsResponse {
	return StatisticsSensorsResponse{
		Today:     s.Today.ToStatisticsSensorResponse(),
		Yesterday: s.Yesterday.ToStatisticsSensorResponse(),
	}
}

// StatisticsSensor represents a single sensor for sale statistics.
type StatisticsSensor struct {
	DateString string
	TotalSales float64
}

func (s StatisticsSensor) ToStatisticsSensorResponse() StatisticsSensorResponse {
	return StatisticsSensorResponse{
		DateString: s.DateString,
		TotalSales: money.FormatRupiah(s.TotalSales),
	}
}
