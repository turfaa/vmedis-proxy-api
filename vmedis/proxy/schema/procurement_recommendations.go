package schema

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

// DrugProcurementRecommendationsResponse represents the drug procurement recommendations API response.
type DrugProcurementRecommendationsResponse struct {
	Recommendations []DrugProcurementRecommendation `json:"recommendations"`
	ComputedAt      time.Time                       `json:"computedAt"`
}

// DrugProcurementRecommendation represents one drug procurement recommendation.
type DrugProcurementRecommendation struct {
	DrugStock    `json:",inline"`
	FromSupplier string  `json:"fromSupplier,omitempty"`
	Procurement  Stock   `json:"procurement"`
	Alternatives []Stock `json:"alternatives,omitempty"`
}

// DrugStock is the stock of a drug.
type DrugStock struct {
	Drug  Drug  `json:"drug"`
	Stock Stock `json:"stock"`
}

// FromClientDrugStock converts DrugStock from client schema to proxy schema.
func FromClientDrugStock(cd client.DrugStock) DrugStock {
	return DrugStock{
		Drug:  FromClientDrug(cd.Drug),
		Stock: FromClientStock(cd.Stock),
	}
}
