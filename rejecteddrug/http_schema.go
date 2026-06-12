package rejecteddrug

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type CreateRejectedDrugRequest struct {
	DrugName string  `json:"drugName" binding:"required"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Reason   string  `json:"reason"`
}

// UpdateRejectedDrugRequest updates a rejected drug entry.
// Nil fields are left unchanged.
type UpdateRejectedDrugRequest struct {
	DrugName        *string                        `json:"drugName"`
	Quantity        *float64                       `json:"quantity"`
	Unit            *string                        `json:"unit"`
	Reason          *string                        `json:"reason"`
	Resolution      *models.RejectedDrugResolution `json:"resolution"`
	ResolutionNotes *string                        `json:"resolutionNotes"`
}

type RejectedDrugResponse struct {
	RejectedDrug RejectedDrug `json:"rejectedDrug"`
}

type RejectedDrugsResponse struct {
	RejectedDrugs []RejectedDrug `json:"rejectedDrugs"`
}

type ResolutionsResponse struct {
	Resolutions []models.RejectedDrugResolution `json:"resolutions"`
}
