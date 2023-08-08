package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// SaleStatisticsResponse represents the sale statistics API response.
type SaleStatisticsResponse struct {
	History []SaleStatistics `json:"history"`
}

// SaleStatistics represents the sale statistics from the beginning of the day until PulledAt.
type SaleStatistics struct {
	PulledAt      time.Time `json:"pulledAt"`
	TotalSales    float64   `json:"totalSales"`
	NumberOfSales int       `json:"numberOfSales"`
}

// FromSalesStatisticsClientSchema converts SaleStatistics from client schema to proxy schema.
func FromSalesStatisticsClientSchema(pulledAt time.Time, salesStatistics client.SalesStatistics) (SaleStatistics, error) {
	totalSales, err := salesStatistics.TotalSalesFloat64()
	if err != nil {
		return SaleStatistics{}, fmt.Errorf("total sales float64: %w", err)
	}

	return SaleStatistics{
		PulledAt:      pulledAt,
		TotalSales:    totalSales,
		NumberOfSales: salesStatistics.NumberOfSales,
	}, nil
}

// FromModelsSaleStatistics converts SaleStatistics from models.SaleStatistics to proxy schema.
func FromModelsSaleStatistics(saleStatistics models.SaleStatistics) SaleStatistics {
	return SaleStatistics{
		PulledAt:      saleStatistics.PulledAt,
		TotalSales:    saleStatistics.TotalSales,
		NumberOfSales: saleStatistics.NumberOfSales,
	}
}

// DrugProcurementRecommendationsResponse represents the drug procurement recommendations API response.
type DrugProcurementRecommendationsResponse struct {
	Recommendations []DrugProcurementRecommendation `json:"recommendations"`
	ComputedAt      time.Time                       `json:"computedAt"`
}

// DrugProcurementRecommendation represents one drug procurement recommendation.
type DrugProcurementRecommendation struct {
	DrugStock    DrugStock `json:",inline"`
	FromSupplier string    `json:"fromSupplier"`
	Procurement  Stock     `json:"procurement"`
}

// DrugStock is the stock of a drug.
type DrugStock struct {
	Drug  Drug  `json:",inline"`
	Stock Stock `json:"stock"`
}

// FromClientDrugStock converts DrugStock from client schema to proxy schema.
func FromClientDrugStock(cd client.DrugStock) DrugStock {
	return DrugStock{
		Drug:  FromClientDrug(cd.Drug),
		Stock: FromClientStock(cd.Stock),
	}
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisCode   string `json:"vmedisCode"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Supplier     string `json:"supplier"`
	MinimumStock Stock  `json:"minimumStock"`
}

// FromClientDrug converts Drug from client schema to proxy schema.
func FromClientDrug(cd client.Drug) Drug {
	return Drug{
		VmedisCode:   cd.VmedisCode,
		Name:         cd.Name,
		Manufacturer: cd.Manufacturer,
		Supplier:     cd.Supplier,
		MinimumStock: FromClientStock(cd.MinimumStock),
	}
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
}

// MarshalText implements encoding.TextMarshaler.
func (s Stock) MarshalText() ([]byte, error) {
	q, err := json.Marshal(s.Quantity)
	if err != nil {
		return nil, fmt.Errorf("marshal quantity: %w", err)
	}

	var b bytes.Buffer
	b.Write(q)
	if s.Unit != "" {
		b.WriteByte(' ')
		b.WriteString(s.Unit)
	}

	return bytes.TrimSpace(b.Bytes()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Stock) UnmarshalText(text []byte) error {
	split := bytes.SplitN(bytes.TrimSpace(text), []byte(" "), 2)

	var q float64
	if err := json.Unmarshal(split[0], &q); err != nil {
		return fmt.Errorf("unmarshal quantity: %w", err)
	}

	s.Quantity = q
	if len(split) > 1 {
		s.Unit = string(split[1])
	}

	return nil
}

// FromClientStock converts Stock from client schema to proxy schema.
func FromClientStock(cs client.Stock) Stock {
	return Stock{
		Unit:     cs.Unit,
		Quantity: cs.Quantity,
	}
}
