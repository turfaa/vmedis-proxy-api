package schema

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
)

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

// FromModelsInvoiceCalculator converts models.InvoiceCalculator to InvoiceCalculator.
func FromModelsInvoiceCalculator(calculator models.InvoiceCalculator) InvoiceCalculator {
	components := make([]InvoiceComponent, len(calculator.Components))
	for i, component := range calculator.Components {
		components[i] = FromModelsInvoiceComponent(component)
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

// FromModelsInvoiceComponent converts models.InvoiceComponent to InvoiceComponent.
func FromModelsInvoiceComponent(component models.InvoiceComponent) InvoiceComponent {
	return InvoiceComponent{
		Name:       component.Name,
		Multiplier: component.Multiplier,
	}
}
