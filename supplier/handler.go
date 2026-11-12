package supplier

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
	service *Service
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{service: service}
}

// GetSuppliers handles requests to get all the suppliers.
func (h *ApiHandler) GetSuppliers(c *gin.Context) {
	suppliers, err := h.service.GetSuppliers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get suppliers: %s", err)})
		return
	}

	c.JSON(http.StatusOK, SuppliersResponse{Suppliers: suppliers})
}

// DumpSuppliers handles requests to start dumping the suppliers from vmedis to
// the DB. Only one dump runs at a time.
func (h *ApiHandler) DumpSuppliers(c *gin.Context) {
	if err := h.service.StartDumpSuppliersFromVmedisToDB(c.Request.Context()); err != nil {
		if errors.Is(err, ErrDumpInProgress) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to start dumping suppliers: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dumping suppliers from vmedis to db"})
}

// GetDumpStatus handles requests to get whether a supplier dump is in progress.
func (h *ApiHandler) GetDumpStatus(c *gin.Context) {
	status, err := h.service.GetDumpStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get supplier dump status: %s", err)})
		return
	}

	c.JSON(http.StatusOK, status)
}
