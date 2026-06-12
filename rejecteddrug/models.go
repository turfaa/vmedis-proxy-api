package rejecteddrug

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type RejectedDrug struct {
	ID              uint                          `json:"id"`
	CreatedAt       time.Time                     `json:"createdAt"`
	UpdatedAt       time.Time                     `json:"updatedAt"`
	DrugName        string                        `json:"drugName"`
	Resolution      models.RejectedDrugResolution `json:"resolution"`
	ResolutionNotes string                        `json:"resolutionNotes"`
	ResolvedAt      *time.Time                    `json:"resolvedAt,omitempty"`
	CreatedBy       string                        `json:"createdBy"`
	ResolvedBy      string                        `json:"resolvedBy"`
}

func FromDBRejectedDrug(rejectedDrug models.RejectedDrug) RejectedDrug {
	return RejectedDrug{
		ID:              rejectedDrug.ID,
		CreatedAt:       rejectedDrug.CreatedAt,
		UpdatedAt:       rejectedDrug.UpdatedAt,
		DrugName:        rejectedDrug.DrugName,
		Resolution:      rejectedDrug.Resolution,
		ResolutionNotes: rejectedDrug.ResolutionNotes,
		ResolvedAt:      rejectedDrug.ResolvedAt,
		CreatedBy:       rejectedDrug.CreatedBy,
		ResolvedBy:      rejectedDrug.ResolvedBy,
	}
}

// ListFilters are the filters that can be applied when listing rejected drugs.
// Zero-valued fields are ignored, so any combination of filters can be used.
type ListFilters struct {
	// Query fuzzy-matches drug name and resolution notes.
	Query           string
	DrugName        string
	ResolutionNotes string
	Resolutions     []models.RejectedDrugResolution
	CreatedBy       string
	ResolvedBy      string
	CreatedFrom     *time.Time
	CreatedUntil    *time.Time
	ResolvedFrom    *time.Time
	ResolvedUntil   *time.Time
}
