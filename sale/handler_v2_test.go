package sale_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/sale"
)

// TestGetLastDrugSales checks that only the sales of the requested drug are
// returned, most recent first, limited to the requested number of rows.
func TestGetLastDrugSales(t *testing.T) {
	router := setupRouter(t, []models.Sale{
		newSale("INV-1", "2024-03-01 08:00", newSaleUnit("INV-1", 1, "D1", 2, "TABLET", 1_000, "Umum", 2_000)),
		newSale("INV-2", "2024-03-03 09:30", newSaleUnit("INV-2", 1, "D1", 1, "TABLET", 1_500, "Resep", 1_500)),
		newSale("INV-3", "2024-03-02 10:15", newSaleUnit("INV-3", 1, "D1", 3, "TABLET", 900, "Reseller", 2_700)),
		// Another drug, must be excluded.
		newSale("INV-4", "2024-03-04 11:00", newSaleUnit("INV-4", 1, "D2", 5, "BOTOL", 20_000, "Umum", 100_000)),
	})

	code, body := do(router, "GET", "/sales/drugs/D1/last?limit=2")
	if code != 200 {
		t.Fatalf("last drug sales: got code %d, body %s", code, body)
	}

	table := unmarshal[cui.Table](t, body)

	if len(table.Rows) != 2 {
		t.Fatalf("last drug sales: expected 2 rows, got %s", body)
	}

	wantRows := [][]string{
		{"2024-03-03 09:30", "INV-2", "1 TABLET", "Rp 1.500 / TABLET", "Rp 1.500", "Resep"},
		{"2024-03-02 10:15", "INV-3", "3 TABLET", "Rp 900 / TABLET", "Rp 2.700", "Reseller"},
	}

	for i, want := range wantRows {
		if len(table.Rows[i].Columns) != len(table.Header) {
			t.Fatalf("last drug sales: row %d has %d columns, header has %d", i, len(table.Rows[i].Columns), len(table.Header))
		}

		for j, wantColumn := range want {
			if table.Rows[i].Columns[j] != wantColumn {
				t.Errorf("last drug sales: row %d column %d: got %q, want %q", i, j, table.Rows[i].Columns[j], wantColumn)
			}
		}
	}
}

// TestGetLastDrugSalesExcludesDeletedSales checks that the units of a deleted
// sale don't show up, the same way the sold drugs reports skip them.
func TestGetLastDrugSalesExcludesDeletedSales(t *testing.T) {
	router, db := setupRouterAndDB(t, []models.Sale{
		newSale("INV-1", "2024-03-01 08:00", newSaleUnit("INV-1", 1, "D1", 2, "TABLET", 1_000, "Umum", 2_000)),
		newSale("INV-2", "2024-03-03 09:30", newSaleUnit("INV-2", 1, "D1", 1, "TABLET", 1_500, "Umum", 1_500)),
	})

	if err := db.Where("invoice_number = ?", "INV-2").Delete(&models.Sale{}).Error; err != nil {
		t.Fatalf("delete sale: %s", err)
	}

	code, body := do(router, "GET", "/sales/drugs/D1/last")
	if code != 200 {
		t.Fatalf("last drug sales: got code %d, body %s", code, body)
	}

	table := unmarshal[cui.Table](t, body)
	if len(table.Rows) != 1 {
		t.Fatalf("last drug sales: expected only the non-deleted sale, got %s", body)
	}
	if table.Rows[0].Columns[1] != "INV-1" {
		t.Fatalf("last drug sales: expected INV-1, got %s", body)
	}
}

// TestGetLastDrugSalesWithoutData checks that a drug that was never sold still
// gets a well-formed, empty table.
func TestGetLastDrugSalesWithoutData(t *testing.T) {
	router := setupRouter(t, nil)

	code, body := do(router, "GET", "/sales/drugs/D1/last")
	if code != 200 {
		t.Fatalf("last drug sales: got code %d, body %s", code, body)
	}

	table := unmarshal[cui.Table](t, body)
	if len(table.Rows) != 0 {
		t.Fatalf("last drug sales: expected no rows, got %s", body)
	}
	if len(table.Header) == 0 {
		t.Fatalf("last drug sales: expected a header, got %s", body)
	}
}

func newSale(invoiceNumber string, soldAt string, units ...models.SaleUnit) models.Sale {
	at, err := time.ParseInLocation("2006-01-02 15:04", soldAt, time.Local)
	if err != nil {
		panic(err)
	}

	return models.Sale{
		InvoiceNumber: invoiceNumber,
		SoldAt:        at,
		SaleUnits:     units,
	}
}

func newSaleUnit(
	invoiceNumber string,
	idInSale int,
	drugCode string,
	amount float64,
	unit string,
	unitPrice float64,
	priceCategory string,
	total float64,
) models.SaleUnit {
	return models.SaleUnit{
		InvoiceNumber: invoiceNumber,
		IDInSale:      idInSale,
		DrugCode:      drugCode,
		DrugName:      "Drug " + drugCode,
		Amount:        amount,
		Unit:          unit,
		UnitPrice:     unitPrice,
		PriceCategory: priceCategory,
		Total:         total,
	}
}

func setupRouter(t *testing.T, sales []models.Sale) *gin.Engine {
	t.Helper()

	router, _ := setupRouterAndDB(t, sales)
	return router
}

func setupRouterAndDB(t *testing.T, sales []models.Sale) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %s", err)
	}
	if err := db.AutoMigrate(&models.Sale{}, &models.SaleUnit{}); err != nil {
		t.Fatalf("migrate database: %s", err)
	}
	if len(sales) > 0 {
		if err := db.Create(&sales).Error; err != nil {
			t.Fatalf("seed sales: %s", err)
		}
	}

	handler := sale.NewApiHandler(sale.NewService(db, nil, nil, nil))

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mirrors the route registration in proxy/api.go, without auth middleware.
	router.GET("/sales/drugs/:drug_code/last", handler.GetLastDrugSales)

	return router, db
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
