package rejecteddrug

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/turfaa/vmedis-proxy-api/auth"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApiHandler struct {
	service *Service
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{service: service}
}

// GetRejectedDrugs returns rejected drugs filtered by query parameters.
// All filters are optional and can be combined:
//   - query: fuzzy-matches drug name, reason, and resolution notes
//   - drug_name, reason, resolution_notes: fuzzy-match their respective fields
//   - resolutions: comma-separated list of resolutions (can also be repeated)
//   - created_by, resolved_by: exact match on user email
//   - date | from + until/to: created-at date range
//   - resolved_from, resolved_until: resolved-at date range
func (h *ApiHandler) GetRejectedDrugs(c *gin.Context) {
	filters, err := extractListFilters(c)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid filters: %s", err)})
		return
	}

	rejectedDrugs, err := h.service.GetRejectedDrugs(c, filters)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get rejected drugs: %s", err)})
		return
	}

	c.JSON(200, RejectedDrugsResponse{RejectedDrugs: rejectedDrugs})
}

func (h *ApiHandler) GetRejectedDrug(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid id: %s", err)})
		return
	}

	rejectedDrug, err := h.service.GetRejectedDrugByID(c, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": fmt.Sprintf("rejected drug %d not found", id)})
			return
		}

		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get rejected drug %d: %s", id, err)})
		return
	}

	c.JSON(200, RejectedDrugResponse{RejectedDrug: rejectedDrug})
}

func (h *ApiHandler) CreateRejectedDrug(c *gin.Context) {
	var request CreateRejectedDrugRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
		return
	}

	rejectedDrug, err := h.service.CreateRejectedDrug(c, request, auth.FromGinContext(c).Email)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create rejected drug: %s", err)})
		return
	}

	c.JSON(201, RejectedDrugResponse{RejectedDrug: rejectedDrug})
}

func (h *ApiHandler) UpdateRejectedDrug(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid id: %s", err)})
		return
	}

	var request UpdateRejectedDrugRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
		return
	}

	rejectedDrug, err := h.service.UpdateRejectedDrug(c, uint(id), request, auth.FromGinContext(c).Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": fmt.Sprintf("rejected drug %d not found", id)})
			return
		}

		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to update rejected drug %d: %s", id, err)})
		return
	}

	c.JSON(200, RejectedDrugResponse{RejectedDrug: rejectedDrug})
}

func (h *ApiHandler) DeleteRejectedDrug(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid id: %s", err)})
		return
	}

	if err := h.service.DeleteRejectedDrug(c, uint(id)); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to delete rejected drug %d: %s", id, err)})
		return
	}

	c.JSON(200, gin.H{"message": "Rejected drug deleted successfully"})
}

func (h *ApiHandler) GetResolutions(c *gin.Context) {
	c.JSON(200, ResolutionsResponse{Resolutions: h.service.GetResolutions()})
}

func extractListFilters(c *gin.Context) (ListFilters, error) {
	filters := ListFilters{
		Query:           c.Query("query"),
		DrugName:        c.Query("drug_name"),
		Reason:          c.Query("reason"),
		ResolutionNotes: c.Query("resolution_notes"),
		CreatedBy:       c.Query("created_by"),
		ResolvedBy:      c.Query("resolved_by"),
	}

	resolutions, err := extractResolutions(c)
	if err != nil {
		return ListFilters{}, err
	}
	filters.Resolutions = resolutions

	// Only filter by creation time when the client sends a date range,
	// because time2.GetTimeRangeFromQuery defaults to today.
	if c.Query("date") != "" || c.Query("from") != "" || c.Query("until") != "" || c.Query("to") != "" {
		from, until, err := time2.GetTimeRangeFromQuery(c)
		if err != nil {
			return ListFilters{}, fmt.Errorf("invalid created-at time range: %w", err)
		}

		filters.CreatedFrom = &from
		filters.CreatedUntil = &until
	}

	if resolvedFrom := c.Query("resolved_from"); resolvedFrom != "" {
		from, err := time2.BeginningOfDate(resolvedFrom)
		if err != nil {
			return ListFilters{}, fmt.Errorf("invalid `resolved_from` query [%s]: %w", resolvedFrom, err)
		}
		filters.ResolvedFrom = &from
	}

	if resolvedUntil := c.Query("resolved_until"); resolvedUntil != "" {
		until, err := time2.EndOfDate(resolvedUntil)
		if err != nil {
			return ListFilters{}, fmt.Errorf("invalid `resolved_until` query [%s]: %w", resolvedUntil, err)
		}
		filters.ResolvedUntil = &until
	}

	return filters, nil
}

func extractResolutions(c *gin.Context) ([]models.RejectedDrugResolution, error) {
	var resolutions []models.RejectedDrugResolution

	values := append(c.QueryArray("resolutions"), c.QueryArray("resolution")...)
	for _, value := range values {
		for _, raw := range strings.Split(value, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}

			resolution := models.RejectedDrugResolution(strings.ToUpper(raw))
			if !resolution.Valid() {
				return nil, fmt.Errorf("invalid resolution: %s", raw)
			}

			resolutions = append(resolutions, resolution)
		}
	}

	return resolutions, nil
}
