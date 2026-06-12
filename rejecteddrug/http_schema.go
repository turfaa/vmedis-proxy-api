package rejecteddrug

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
)

type CreateRejectedDrugRequest struct {
	DrugName string `json:"drugName" binding:"required"`
}

// UpdateRejectedDrugRequest updates a rejected drug entry.
// Nil fields are left unchanged.
type UpdateRejectedDrugRequest struct {
	DrugName        *string                        `json:"drugName"`
	Resolution      *models.RejectedDrugResolution `json:"resolution"`
	ResolutionNotes *string                        `json:"resolutionNotes"`
}

type RejectedDrugResponse struct {
	RejectedDrug RejectedDrug `json:"rejectedDrug"`
}
