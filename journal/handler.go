package journal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/auth"
)

type ApiHandler struct {
	service *Service
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{service: service}
}

// Create handles requests to create a journal entry in vmedis. The
// Idempotency-Key header is required.
func (h *ApiHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
		return
	}

	res, err := h.service.Create(c.Request.Context(), c.GetHeader("Idempotency-Key"), req, auth.FromGinContext(c).Email)
	if err != nil {
		h.writeError(c, "failed to create journal entry", err)
		return
	}

	status := http.StatusCreated
	if res.Replayed {
		status = http.StatusOK
	}

	c.JSON(status, res)
}

// Find handles requests to list the journal entries of a date (?date=), whose
// description contains ?q=.
func (h *ApiHandler) Find(c *gin.Context) {
	res, err := h.service.Find(c.Request.Context(), c.Query("date"), c.Query("q"))
	if err != nil {
		h.writeError(c, "failed to find journal entries", err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ApiHandler) writeError(c *gin.Context, message string, err error) {
	var validationErr *ValidationError

	switch {
	case errors.As(err, &validationErr):
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})

	case errors.Is(err, ErrInProgress), errors.Is(err, ErrKeyReused):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, ErrVmedis):
		// outcomeUnknown tells the client the entry may exist in vmedis, so it
		// must not treat the request as failed: a retry with the same key
		// returns the entry if it does.
		c.JSON(http.StatusBadGateway, gin.H{
			"error":          fmt.Sprintf("%s: %s", message, err),
			"outcomeUnknown": errors.Is(err, ErrOutcomeUnknown),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", message, err)})
	}
}
