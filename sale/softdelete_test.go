package sale

import (
	"context"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

// TestSaleSoftDeleteRoundTrip checks that a soft-deleted sale disappears from
// the reads, and that dumping it again from Vmedis brings it and its units back
// to life in place, rather than leaving a second sale with the same invoice
// number behind.
func TestSaleSoftDeleteRoundTrip(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	d := NewDatabase(db)

	soldAt := time.Now().Truncate(time.Second)
	from, to := soldAt.Add(-time.Hour), soldAt.Add(time.Hour)

	const invoiceNumber = "PJ2607120001"

	dump := func(total float64) {
		t.Helper()

		if err := d.UpsertVmedisSales(ctx, []vmedisv1.Sale{{
			ID:            1,
			Date:          vmedisv1.Time{Time: soldAt},
			InvoiceNumber: invoiceNumber,
			Total:         total,
			SaleUnits: []vmedisv1.SaleUnit{
				{IDInSale: 1, DrugCode: "D1", DrugName: "Drug One", Amount: 2, Unit: "tablet", Total: total},
			},
		}}); err != nil {
			t.Fatalf("dump sale: %v", err)
		}
	}

	// assertState checks how many sales the reads see, and how many rows are
	// actually stored, so that a "revived" sale that is really a duplicate is
	// caught.
	assertState := func(stage string, wantVisible int, wantStoredSales, wantStoredUnits, wantVisibleUnits int64) {
		t.Helper()

		sales, err := d.GetSalesBetweenTime(ctx, from, to)
		if err != nil {
			t.Fatalf("%s: get sales: %v", stage, err)
		}

		var storedSales, storedUnits, visibleUnits int64
		db.Unscoped().Model(&models.Sale{}).Count(&storedSales)
		db.Unscoped().Model(&models.SaleUnit{}).Count(&storedUnits)
		db.Model(&models.SaleUnit{}).Count(&visibleUnits)

		if len(sales) != wantVisible {
			t.Fatalf("%s: %d visible sales, want %d", stage, len(sales), wantVisible)
		}
		if storedSales != wantStoredSales {
			t.Fatalf("%s: %d stored sale rows, want %d", stage, storedSales, wantStoredSales)
		}
		if storedUnits != wantStoredUnits {
			t.Fatalf("%s: %d stored sale unit rows, want %d", stage, storedUnits, wantStoredUnits)
		}
		if visibleUnits != wantVisibleUnits {
			t.Fatalf("%s: %d visible sale units, want %d", stage, visibleUnits, wantVisibleUnits)
		}
	}

	dump(10)
	assertState("after first dump", 1, 1, 1, 1)

	if err := d.DeleteSaleByInvoiceNumber(ctx, invoiceNumber); err != nil {
		t.Fatalf("delete sale: %v", err)
	}
	assertState("after soft delete", 0, 1, 1, 0)

	dump(30)
	assertState("after re-dump", 1, 1, 1, 1)

	sales, err := d.GetSalesBetweenTime(ctx, from, to)
	if err != nil {
		t.Fatalf("get sales: %v", err)
	}
	if sales[0].Total != 30 {
		t.Errorf("revived sale total = %v, want 30", sales[0].Total)
	}
	if len(sales[0].SaleUnits) != 1 {
		t.Errorf("revived sale has %d units, want 1", len(sales[0].SaleUnits))
	}
}
