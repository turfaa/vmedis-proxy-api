package procurement

import (
	"time"

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
