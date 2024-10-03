package procurement

import (
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/drug"
)

type DumpProcurementsRequest struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func (r DumpProcurementsRequest) ExtractDates() (time.Time, time.Time, error) {
	if r.StartDate == "" || r.EndDate == "" {
		return time.Now().AddDate(0, 0, -14), time.Now(), nil
	}

	startDate, err := time.Parse(time.DateOnly, r.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse start date '%s': %w", r.StartDate, err)
	}

	endDate, err := time.Parse(time.DateOnly, r.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse end date '%s': %w", r.EndDate, err)
	}

	return startDate, endDate, nil
}

type RecommendationsResponse struct {
	Recommendations []Recommendation `json:"recommendations"`
	ComputedAt      time.Time        `json:"computedAt"`
}

type Recommendation struct {
	DrugStock    drug.WithStock `json:",inline"`
	FromSupplier string         `json:"fromSupplier,omitempty"`
	Procurement  drug.Stock     `json:"procurement"`
	Alternatives []drug.Stock   `json:"alternatives,omitempty"`
}

// InvoiceCalculatorsResponse is the response schema for the invoice calculators API.
type InvoiceCalculatorsResponse struct {
	Calculators []InvoiceCalculator `json:"calculators"`
}

// InvoiceCalculator represents an invoice calculator for a supplier.
type InvoiceCalculator struct {
	Supplier    string             `json:"supplier"`
	ShouldRound bool               `json:"shouldRound"`
	Components  []InvoiceComponent `json:"components"`
}

// FromDBInvoiceCalculator converts models.InvoiceCalculator to InvoiceCalculator.
func FromDBInvoiceCalculator(calculator models.InvoiceCalculator) InvoiceCalculator {
	components := make([]InvoiceComponent, len(calculator.Components))
	for i, component := range calculator.Components {
		components[i] = FromDBInvoiceComponent(component)
	}

	return InvoiceCalculator{
		Supplier:    calculator.Supplier,
		ShouldRound: calculator.ShouldRound,
		Components:  components,
	}
}

// InvoiceComponent represents an invoice component.
type InvoiceComponent struct {
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
}

// FromDBInvoiceComponent converts models.InvoiceComponent to InvoiceComponent.
func FromDBInvoiceComponent(component models.InvoiceComponent) InvoiceComponent {
	return InvoiceComponent{
		Name:       component.Name,
		Multiplier: component.Multiplier,
	}
}

type AggregatedProcurement struct {
	DrugName string  `json:"drugName"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type LastDrugProcurementsRequest struct {
	DrugCode string `json:"drugCode" uri:"drug_code"`
	Limit    int    `json:"limit" form:"limit"`
}

type DrugProcurement struct {
	CreatedAt      time.Time `json:"createdAt"`
	DrugCode       string    `json:"drugCode"`
	DrugName       string    `json:"drugName"`
	Amount         float64   `json:"amount"`
	Unit           string    `json:"unit"`
	TotalUnitPrice float64   `json:"totalUnitPrice"`

	InvoiceNumber string    `json:"invoiceNumber"`
	InvoiceDate   time.Time `json:"invoiceDate"`
	Supplier      string    `json:"supplier"`
}
