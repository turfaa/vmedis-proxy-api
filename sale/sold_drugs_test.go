package sale

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/drug"
)

// fakeDrugsGetter returns a drug for every code it is asked for, and records
// the codes.
type fakeDrugsGetter struct {
	calls [][]string
}

func (g *fakeDrugsGetter) GetDrugsByVmedisCodes(_ context.Context, codes []string) ([]drug.Drug, error) {
	g.calls = append(g.calls, codes)

	drugs := make([]drug.Drug, len(codes))
	for i, code := range codes {
		drugs[i] = drug.Drug{VmedisCode: code, Name: "Drug " + code}
	}
	return drugs, nil
}

func TestGetSoldDrugsBetweenTime(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	day := time.Date(2026, 11, 12, 10, 0, 0, 0, time.Local)
	from, to := day.Add(-time.Hour), day.Add(time.Hour)

	sales := []models.Sale{
		{InvoiceNumber: "PJ1", SoldAt: day, SaleUnits: []models.SaleUnit{
			{IDInSale: 1, DrugCode: "A", Total: 100},
			{IDInSale: 2, DrugCode: "B", Total: 50},
		}},
		{InvoiceNumber: "PJ2", SoldAt: day, SaleUnits: []models.SaleUnit{
			{IDInSale: 1, DrugCode: "A", Total: 200},
			// Soft-deleted below, so it is not counted.
			{IDInSale: 2, DrugCode: "C", Total: 1000},
		}},
		// Outside the time range.
		{InvoiceNumber: "PJ3", SoldAt: day.Add(-2 * time.Hour), SaleUnits: []models.SaleUnit{
			{IDInSale: 1, DrugCode: "A", Total: 999},
		}},
		// Soft-deleted below, so none of its units are counted.
		{InvoiceNumber: "PJ4", SoldAt: day, SaleUnits: []models.SaleUnit{
			{IDInSale: 1, DrugCode: "B", Total: 999},
		}},
	}
	if err := db.Create(&sales).Error; err != nil {
		t.Fatalf("create sales: %v", err)
	}
	if err := db.Where("invoice_number = ? AND drug_code = ?", "PJ2", "C").Delete(&models.SaleUnit{}).Error; err != nil {
		t.Fatalf("soft-delete sale unit: %v", err)
	}
	if err := db.Where("invoice_number = ?", "PJ4").Delete(&models.Sale{}).Error; err != nil {
		t.Fatalf("soft-delete sale: %v", err)
	}

	drugs := &fakeDrugsGetter{}
	s := NewService(db, nil, drugs, nil)

	got, err := s.GetSoldDrugsBetweenTime(ctx, from, to)
	if err != nil {
		t.Fatalf("GetSoldDrugsBetweenTime: %v", err)
	}

	want := []SoldDrug{
		{Drug: drug.Drug{VmedisCode: "A", Name: "Drug A"}, Occurrences: 2, TotalAmount: 300},
		{Drug: drug.Drug{VmedisCode: "B", Name: "Drug B"}, Occurrences: 1, TotalAmount: 50},
	}
	if !slices.EqualFunc(got, want, func(a, b SoldDrug) bool {
		return a.Drug.VmedisCode == b.Drug.VmedisCode && a.Drug.Name == b.Drug.Name && a.Occurrences == b.Occurrences && a.TotalAmount == b.TotalAmount
	}) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if len(drugs.calls) != 1 || len(drugs.calls[0]) != 2 {
		t.Errorf("got drug lookups %v, want one lookup of the 2 distinct codes", drugs.calls)
	}
}

// TestGetSoldDrugsBetweenTimeManySaleUnits checks that a time range with more
// sales and sale units than a query can have parameters still works.
func TestGetSoldDrugsBetweenTimeManySaleUnits(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	const (
		numSales     = 35_000
		unitsPerSale = 2
		numDrugs     = 1_500
	)

	day := time.Date(2026, 11, 12, 10, 0, 0, 0, time.Local)

	// Generate the rows in SQL; creating them one by one through GORM takes
	// minutes under the race detector.
	if err := db.Exec(`
		WITH RECURSIVE seq(i) AS (SELECT 0 UNION ALL SELECT i + 1 FROM seq WHERE i + 1 < ?)
		INSERT INTO sales (invoice_number, sold_at) SELECT 'PJ' || i, ? FROM seq`,
		numSales, day,
	).Error; err != nil {
		t.Fatalf("create sales: %v", err)
	}
	if err := db.Exec(`
		WITH RECURSIVE seq(i) AS (SELECT 0 UNION ALL SELECT i + 1 FROM seq WHERE i + 1 < ?)
		INSERT INTO sale_units (invoice_number, id_in_sale, drug_code, total)
		SELECT 'PJ' || (i / ?), (i % ?) + 1, 'D' || (i % ?), 10 FROM seq`,
		numSales*unitsPerSale, unitsPerSale, unitsPerSale, numDrugs,
	).Error; err != nil {
		t.Fatalf("create sale units: %v", err)
	}

	drugs := &fakeDrugsGetter{}
	s := NewService(db, nil, drugs, nil)

	got, err := s.GetSoldDrugsBetweenTime(ctx, day.Add(-time.Hour), day.Add(time.Hour))
	if err != nil {
		t.Fatalf("GetSoldDrugsBetweenTime: %v", err)
	}

	if len(got) != numDrugs {
		t.Errorf("got %d sold drugs, want %d", len(got), numDrugs)
	}

	var occurrences int
	for _, d := range got {
		occurrences += d.Occurrences
	}
	if occurrences != numSales*unitsPerSale {
		t.Errorf("got %d occurrences in total, want %d", occurrences, numSales*unitsPerSale)
	}
}
