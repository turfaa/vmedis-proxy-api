package purchaseorder

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/auth"
	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"
)

type ApiHandler struct {
	service *Service
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{service: service}
}

// GetOptions handles requests to get what the create dialog needs: the
// templates, the signer and the vmedis supplier page.
func (h *ApiHandler) GetOptions(c *gin.Context) {
	options, err := h.service.GetOptions()
	if err != nil {
		h.writeError(c, "failed to get purchase order options", err)
		return
	}

	c.JSON(http.StatusOK, options)
}

// Create handles requests to create a purchase order.
func (h *ApiHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
		return
	}

	res, err := h.service.Create(c.Request.Context(), req, auth.FromGinContext(c).Email)
	if err != nil {
		h.writeError(c, "failed to create purchase order", err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

// List handles requests to list purchase orders.
func (h *ApiHandler) List(c *gin.Context) {
	filter := ListFilter{
		Query:      c.Query("q"),
		TemplateID: c.Query("templateId"),
	}

	var err error
	if from := c.Query("from"); from != "" {
		if filter.From, err = time2.BeginningOfDate(from); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid from: %s", err)})
			return
		}
	}

	if to := c.Query("to"); to != "" {
		if filter.Until, err = time2.EndOfDate(to); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid to: %s", err)})
			return
		}
	}

	for param, dst := range map[string]*int{
		"supplierVmedisId":    &filter.SupplierVmedisID,
		"createdByEmployeeId": &filter.CreatedByEmployeeID,
		"page":                &filter.Page,
		"pageSize":            &filter.PageSize,
	} {
		value := c.Query(param)
		if value == "" {
			continue
		}

		if *dst, err = strconv.Atoi(value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s: %s", param, err)})
			return
		}
	}

	res, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.writeError(c, "failed to list purchase orders", err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// ListCreators handles requests to list the employees who have created
// purchase orders.
func (h *ApiHandler) ListCreators(c *gin.Context) {
	res, err := h.service.ListCreators(c.Request.Context())
	if err != nil {
		h.writeError(c, "failed to list purchase order creators", err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Get handles requests to get a purchase order with its items.
func (h *ApiHandler) Get(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	res, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, "failed to get purchase order", err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetPDF handles requests to download a purchase order as a PDF.
func (h *ApiHandler) GetPDF(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	pdf, err := h.service.RenderPDF(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, "failed to render purchase order", err)
		return
	}

	c.Header("Content-Disposition", contentDisposition(pdf.Filename))
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/pdf", pdf.Content)
}

func (h *ApiHandler) parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase order id"})
		return 0, false
	}

	return uint(id), true
}

func (h *ApiHandler) writeError(c *gin.Context, message string, err error) {
	var validationErr *ValidationError

	switch {
	case errors.As(err, &validationErr):
		body := gin.H{"error": validationErr.Message}
		if len(validationErr.InvalidItems) > 0 {
			body["invalidItems"] = validationErr.InvalidItems
		}
		c.JSON(http.StatusBadRequest, body)

	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s: %s", message, err)})

	case errors.Is(err, ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", message, err)})
	}
}

// contentDisposition returns an attachment Content-Disposition with both a
// plain ASCII filename and the UTF-8 one (RFC 6266).
func contentDisposition(filename string) string {
	return fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		asciiFilename(filename),
		url.PathEscape(filename),
	)
}
