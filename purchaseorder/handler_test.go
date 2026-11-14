package purchaseorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewApiHandler(service)
	r := gin.New()
	g := r.Group("/purchase-orders")
	g.GET("/options", h.GetOptions)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.GET("/:id/pdf", h.GetPDF)

	return r
}

func doRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHandlerStatuses(t *testing.T) {
	env := newTestEnv(t)
	r := newTestRouter(env.service)

	invalid := createRequest(1, CreateItem{DrugName: "A", Quantity: ""})
	rec := doRequest(r, http.MethodPost, "/purchase-orders", invalid)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"invalidItems":[0]`) {
		t.Errorf("invalid items: %d %s", rec.Code, rec.Body)
	}

	if rec := doRequest(r, http.MethodPost, "/purchase-orders", createRequest(99)); rec.Code != http.StatusNotFound {
		t.Errorf("unknown supplier: %d %s", rec.Code, rec.Body)
	}

	rec = doRequest(r, http.MethodPost, "/purchase-orders", createRequest(1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}

	var created CreateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	rec = doRequest(r, http.MethodGet, fmt.Sprintf("/purchase-orders/%d/pdf", created.ID), nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/pdf" {
		t.Errorf("pdf: %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	if want := `attachment; filename="2026-11-14 - PT Supplier Satu - Surat Pesanan Reguler - 001-SP-REG-AAF-XI-2026.pdf"`; !strings.HasPrefix(rec.Header().Get("Content-Disposition"), want) {
		t.Errorf("content disposition = %q", rec.Header().Get("Content-Disposition"))
	}

	for path, want := range map[string]int{
		"/purchase-orders/abc":                    http.StatusBadRequest,
		"/purchase-orders/999":                    http.StatusNotFound,
		"/purchase-orders/999/pdf":                http.StatusNotFound,
		"/purchase-orders?from=14-11-2026":        http.StatusBadRequest,
		"/purchase-orders?page=x":                 http.StatusBadRequest,
		"/purchase-orders?from=2026-11-01&q=para": http.StatusOK,
	} {
		if rec := doRequest(r, http.MethodGet, path, nil); rec.Code != want {
			t.Errorf("GET %s: %d, want %d (%s)", path, rec.Code, want, rec.Body)
		}
	}

	env.service.config = Config{}
	if rec := doRequest(r, http.MethodGet, "/purchase-orders/options", nil); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("options when not configured: %d", rec.Code)
	}
}
