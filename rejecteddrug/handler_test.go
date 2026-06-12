package rejecteddrug_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/rejecteddrug"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestRejectedDrugJourney walks through the whole rejected drug management
// journey from the frontend's point of view: record a rejected drug, browse
// the list, open the detail, load the prefilled update form, resolve the
// entry, and delete it.
func TestRejectedDrugJourney(t *testing.T) {
	router := setupRouter(t)

	do := func(method, path, body string) (int, string) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}

	code, body := do("GET", "/rejected-drugs/form", "")
	if code != 200 {
		t.Fatalf("get create form: got code %d, body %s", code, body)
	}
	createForm := unmarshal[cui.Form](t, body)
	if len(createForm.Fields) != 1 || createForm.Fields[0].ID != "drugName" {
		t.Fatalf("unexpected create form: %s", body)
	}

	code, body = do("POST", "/rejected-drugs", `{"drugName": "Paracetamol 500mg"}`)
	if code != 201 {
		t.Fatalf("create: got code %d, body %s", code, body)
	}
	created := unmarshal[rejecteddrug.RejectedDrugResponse](t, body)
	if created.RejectedDrug.Resolution != models.RejectedDrugResolutionUnresolved {
		t.Fatalf("create: got resolution %s, want UNRESOLVED", created.RejectedDrug.Resolution)
	}

	code, body = do("GET", "/rejected-drugs", "")
	if code != 200 {
		t.Fatalf("list: got code %d, body %s", code, body)
	}
	list := unmarshal[cui.Table](t, body)
	if len(list.Rows) != 1 || list.Rows[0].ID != "1" {
		t.Fatalf("list: expected one row with ID 1, got %s", body)
	}
	if len(list.Rows[0].Columns) != len(list.Header) {
		t.Fatalf("list: row has %d columns, header has %d", len(list.Rows[0].Columns), len(list.Header))
	}

	code, body = do("GET", "/rejected-drugs/1", "")
	if code != 200 {
		t.Fatalf("detail: got code %d, body %s", code, body)
	}
	detail := unmarshal[cui.Table](t, body)
	if len(detail.Rows) == 0 || len(detail.Rows[0].Columns) != 2 {
		t.Fatalf("detail: expected key-value rows, got %s", body)
	}

	code, body = do("GET", "/rejected-drugs/1/form", "")
	if code != 200 {
		t.Fatalf("get update form: got code %d, body %s", code, body)
	}
	updateForm := unmarshal[cui.Form](t, body)
	fields := make(map[string]cui.Field, len(updateForm.Fields))
	for _, field := range updateForm.Fields {
		fields[field.ID] = field
	}
	if fields["drugName"].Value != "Paracetamol 500mg" {
		t.Fatalf("update form: drugName not prefilled: %s", body)
	}
	if fields["resolution"].Value != "UNRESOLVED" || len(fields["resolution"].Options) == 0 {
		t.Fatalf("update form: resolution select not prefilled with options: %s", body)
	}

	code, body = do("GET", "/rejected-drugs/resolutions", "")
	if code != 200 {
		t.Fatalf("resolutions: got code %d, body %s", code, body)
	}
	resolutions := unmarshal[cui.Options](t, body)
	if len(resolutions.Options) != len(models.AllRejectedDrugResolutions()) {
		t.Fatalf("resolutions: expected %d options, got %s", len(models.AllRejectedDrugResolutions()), body)
	}

	code, body = do("PATCH", "/rejected-drugs/1", `{"resolution": "ORDERED", "resolutionNotes": "Dipesan ke PBF"}`)
	if code != 200 {
		t.Fatalf("resolve: got code %d, body %s", code, body)
	}
	resolved := unmarshal[rejecteddrug.RejectedDrugResponse](t, body)
	if resolved.RejectedDrug.Resolution != models.RejectedDrugResolutionOrdered || resolved.RejectedDrug.ResolvedAt == nil {
		t.Fatalf("resolve: entry not marked as resolved: %s", body)
	}

	code, body = do("GET", "/rejected-drugs/1", "")
	if code != 200 {
		t.Fatalf("detail after resolve: got code %d, body %s", code, body)
	}
	if !strings.Contains(body, "Sudah Dipesan") || !strings.Contains(body, "Dipesan ke PBF") {
		t.Fatalf("detail after resolve: missing resolution label or notes: %s", body)
	}

	code, body = do("DELETE", "/rejected-drugs/1", "")
	if code != 200 {
		t.Fatalf("delete: got code %d, body %s", code, body)
	}

	code, body = do("GET", "/rejected-drugs/1", "")
	if code != 404 {
		t.Fatalf("detail after delete: got code %d, body %s", code, body)
	}
}

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %s", err)
	}
	if err := db.AutoMigrate(&models.RejectedDrug{}); err != nil {
		t.Fatalf("migrate database: %s", err)
	}

	handler := rejecteddrug.NewApiHandler(rejecteddrug.NewService(db))

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mirrors the route registration in proxy/api.go, without auth middleware.
	rejectedDrugs := router.Group("/rejected-drugs")
	{
		rejectedDrugs.GET("", handler.GetRejectedDrugs)
		rejectedDrugs.POST("", handler.CreateRejectedDrug)
		rejectedDrugs.GET("/resolutions", handler.GetResolutions)
		rejectedDrugs.GET("/form", handler.GetCreateRejectedDrugForm)
		rejectedDrugs.GET("/:id", handler.GetRejectedDrug)
		rejectedDrugs.GET("/:id/form", handler.GetUpdateRejectedDrugForm)
		rejectedDrugs.PATCH("/:id", handler.UpdateRejectedDrug)
		rejectedDrugs.DELETE("/:id", handler.DeleteRejectedDrug)
	}

	return router
}

func unmarshal[T any](t *testing.T, body string) T {
	t.Helper()

	var value T
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		t.Fatalf("unmarshal %T from %s: %s", value, body, err)
	}

	return value
}
