package procurement_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/procurement"
)

// TestGetSupplierProcurementRecaps checks that the recap groups the procurements
// of the requested invoice date range per supplier, sorts them by total amount
// descending, and sums everything up in the footer.
func TestGetSupplierProcurementRecaps(t *testing.T) {
	router := setupRouter(t, []models.Procurement{
		newProcurement("INV-1", "2024-03-01", "PBF Alpha", 1_000_000),
		newProcurement("INV-2", "2024-03-05", "PBF Alpha", 500_000),
		newProcurement("INV-3", "2024-03-03", "PBF Beta", 2_000_000),
		newProcurement("INV-4", "2024-03-31", "PBF Gamma", 250_000),
		// Outside of the requested range, must be excluded.
		newProcurement("INV-5", "2024-02-29", "PBF Delta", 9_000_000),
		newProcurement("INV-6", "2024-04-01", "PBF Delta", 9_000_000),
	})

	code, body := do(router, "GET", "/procurements/suppliers/recap?from=2024-03-01&until=2024-03-31")
	if code != 200 {
		t.Fatalf("recap: got code %d, body %s", code, body)
	}

	table := unmarshal[cui.Table](t, body)

	if len(table.Rows) != 3 {
		t.Fatalf("recap: expected 3 suppliers, got %s", body)
	}

	wantRows := [][]string{
		{"1", "PBF Beta", "1 Faktur", "Rp 2.000.000"},
		{"2", "PBF Alpha", "2 Faktur", "Rp 1.500.000"},
		{"3", "PBF Gamma", "1 Faktur", "Rp 250.000"},
	}

	for i, want := range wantRows {
		if len(table.Rows[i].Columns) != len(table.Header) {
			t.Fatalf("recap: row %d has %d columns, header has %d", i, len(table.Rows[i].Columns), len(table.Header))
		}

		for j, wantColumn := range want {
			if table.Rows[i].Columns[j] != wantColumn {
				t.Errorf("recap: row %d column %d: got %q, want %q", i, j, table.Rows[i].Columns[j], wantColumn)
			}
		}
	}

	wantFooter := []string{"", "Total", "4 Faktur", "Rp 3.750.000"}
	if len(table.Footer) != len(wantFooter) {
		t.Fatalf("recap: footer has %d columns, want %d: %s", len(table.Footer), len(wantFooter), body)
	}
	for i, want := range wantFooter {
		if table.Footer[i] != want {
			t.Errorf("recap: footer column %d: got %q, want %q", i, table.Footer[i], want)
		}
	}
}

// TestGetSupplierProcurementRecapsWithoutData checks that an empty range still
// returns a well-formed table with a zeroed footer.
func TestGetSupplierProcurementRecapsWithoutData(t *testing.T) {
	router := setupRouter(t, nil)

	code, body := do(router, "GET", "/procurements/suppliers/recap?from=2024-03-01&until=2024-03-31")
	if code != 200 {
		t.Fatalf("recap: got code %d, body %s", code, body)
	}

	table := unmarshal[cui.Table](t, body)
	if len(table.Rows) != 0 {
		t.Fatalf("recap: expected no rows, got %s", body)
	}
	if len(table.Footer) != len(table.Header) {
		t.Fatalf("recap: footer has %d columns, header has %d: %s", len(table.Footer), len(table.Header), body)
	}
	if table.Footer[len(table.Footer)-1] != "Rp 0" {
		t.Fatalf("recap: expected zero total in footer, got %s", body)
	}
}

func newProcurement(invoiceNumber string, invoiceDate string, supplier string, total float64) models.Procurement {
	date, err := time.ParseInLocation(time.DateOnly, invoiceDate, time.Local)
	if err != nil {
		panic(err)
	}

	return models.Procurement{
		InvoiceNumber: invoiceNumber,
		InvoiceDate:   datatypes.Date(date),
		Supplier:      supplier,
		Total:         total,
	}
}

func setupRouter(t *testing.T, procurements []models.Procurement) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %s", err)
	}
	if err := db.AutoMigrate(&models.Procurement{}, &models.ProcurementUnit{}); err != nil {
		t.Fatalf("migrate database: %s", err)
	}
	if len(procurements) > 0 {
		if err := db.Create(&procurements).Error; err != nil {
			t.Fatalf("seed procurements: %s", err)
		}
	}

	handler := procurement.NewApiHandler(procurement.NewService(db, nil, nil, nil, nil))

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mirrors the route registration in proxy/api.go, without auth middleware.
	router.GET("/procurements/suppliers/recap", handler.GetSupplierProcurementRecaps)

	return router
}

func do(router *gin.Engine, method string, path string) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Code, w.Body.String()
}

func unmarshal[T any](t *testing.T, body string) T {
	t.Helper()

	var value T
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		t.Fatalf("unmarshal %s: %s", body, err)
	}

	return value
}
