package procurement

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/drug"
)

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
