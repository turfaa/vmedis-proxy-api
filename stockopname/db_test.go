package stockopname

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func TestVmedisStockOpnameIDs(t *testing.T) {
	oct30 := time.Date(2026, 10, 30, 20, 0, 0, 0, time.Local)
	oct31 := time.Date(2026, 10, 31, 8, 0, 0, 0, time.Local)

	got := VmedisStockOpnameIDs([]vmedisv2.StockOpname{
		{ID: 1, Code: "SOP1", Time: oct30},
		{ID: 2, Code: "SOP2", Time: oct30},
		{ID: 3, Code: "SOP2", Time: oct30},
		{ID: 4, Code: "SOP2", Time: oct30},
		// A code repeated on another day starts over, like the per-day v1
		// dumps did.
		{ID: 5, Code: "SOP2", Time: oct31},
		{ID: 6, Code: "SOP2", Time: oct31},
		{ID: 7, Time: oct31},
	})

	want := []string{
		"SOP1",
		"SOP2",
		"SOP2-augmented-because-duplicate-2",
		"SOP2-augmented-because-duplicate-3",
		"SOP2",
		"SOP2-augmented-because-duplicate-2",
		"sopid-7",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestUpsertVmedisStockOpnames checks that dumping a day again updates the
// rows in place, keyed by the stock opname code.
func TestUpsertVmedisStockOpnames(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	d := NewDatabase(db)

	at := time.Date(2026, 10, 2, 9, 13, 36, 0, time.Local)
	dump := func(notes string) {
		t.Helper()

		err := d.UpsertVmedisStockOpnames(ctx, []vmedisv2.StockOpname{
			{
				ID:                  99310,
				Code:                "SOP202610020002",
				Time:                at,
				DrugCode:            "A00349",
				DrugName:            "POLYCROL FORTE GEL 180ML",
				Unit:                "Botol",
				BatchCode:           "1348",
				InitialQuantity:     2,
				RealQuantity:        0,
				QuantityDifference:  -2,
				HPPDifference:       -71706,
				SalePriceDifference: -81000,
				Notes:               notes,
			},
			{ID: 99311, Code: "SOP202610020002", Time: at, DrugCode: "OBT1"},
		})
		if err != nil {
			t.Fatalf("UpsertVmedisStockOpnames: %v", err)
		}
	}

	dump("ed oktober 2026")
	dump("ed november 2026")

	var rows []models.StockOpname
	if err := db.Order("vmedis_id").Find(&rows).Error; err != nil {
		t.Fatalf("find stock opnames: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}

	got := rows[0]
	if got.VmedisID != "SOP202610020002" || got.DrugCode != "A00349" || got.BatchCode != "1348" ||
		got.InitialQuantity != 2 || got.RealQuantity != 0 || got.QuantityDifference != -2 ||
		got.HPPDifference != -71706 || got.SalePriceDifference != -81000 || got.Notes != "ed november 2026" {
		t.Errorf("got %+v", got)
	}
	if time.Time(got.Date).Format(time.DateOnly) != "2026-10-02" {
		t.Errorf("got date %v, want 2026-10-02", time.Time(got.Date))
	}

	if rows[1].VmedisID != "SOP202610020002-augmented-because-duplicate-2" || rows[1].DrugCode != "OBT1" {
		t.Errorf("got %+v", rows[1])
	}
}
