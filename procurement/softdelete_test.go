package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

// TestProcurementSoftDeleteRoundTrip checks that a soft-deleted procurement
// disappears from the reads, and that dumping it again from Vmedis brings it
// and its units back to life in place, rather than leaving a second procurement
// with the same invoice number behind.
func TestProcurementSoftDeleteRoundTrip(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	d := NewDatabase(db)

	day := time.Now().Truncate(24 * time.Hour)
	from, to := day.Add(-24*time.Hour), day.Add(24*time.Hour)

	const invoiceNumber = "OBT2607120001"

	dump := func(total float64) {
		t.Helper()

		if err := d.UpsertVmedisProcurements(ctx, []vmedisv1.Procurement{{
			Date:          vmedisv1.Date{Time: day},
			InputTime:     vmedisv1.Time{Time: day},
			InvoiceNumber: invoiceNumber,
			Supplier:      "Supplier A",
			Total:         total,
			ProcurementUnits: []vmedisv1.ProcurementUnit{
				{IDInProcurement: 1, DrugCode: "D1", DrugName: "Drug One", Amount: 5, Unit: "box", Total: total},
			},
		}}); err != nil {
			t.Fatalf("dump procurement: %v", err)
		}
	}

	// assertState checks how many procurements the reads see, and how many rows
	// are actually stored, so that a "revived" procurement that is really a
	// duplicate is caught.
	assertState := func(stage string, wantVisible int, wantStored, wantStoredUnits, wantVisibleUnits int64) {
		t.Helper()

		recaps, err := d.GetSupplierProcurementRecapsBetweenTime(ctx, from, to)
		if err != nil {
			t.Fatalf("%s: get supplier recaps: %v", stage, err)
		}

		var stored, storedUnits, visibleUnits int64
		db.Unscoped().Model(&models.Procurement{}).Count(&stored)
		db.Unscoped().Model(&models.ProcurementUnit{}).Count(&storedUnits)
		db.Model(&models.ProcurementUnit{}).Count(&visibleUnits)

		if len(recaps) != wantVisible {
			t.Fatalf("%s: %d supplier recaps, want %d", stage, len(recaps), wantVisible)
		}
		if stored != wantStored {
			t.Fatalf("%s: %d stored procurement rows, want %d", stage, stored, wantStored)
		}
		if storedUnits != wantStoredUnits {
			t.Fatalf("%s: %d stored procurement unit rows, want %d", stage, storedUnits, wantStoredUnits)
		}
		if visibleUnits != wantVisibleUnits {
			t.Fatalf("%s: %d visible procurement units, want %d", stage, visibleUnits, wantVisibleUnits)
		}
	}

	dump(100)
	assertState("after first dump", 1, 1, 1, 1)

	if err := d.DeleteProcurementByInvoiceNumber(ctx, invoiceNumber); err != nil {
		t.Fatalf("delete procurement: %v", err)
	}
	assertState("after soft delete", 0, 1, 1, 0)

	dump(300)
	assertState("after re-dump", 1, 1, 1, 1)

	recaps, err := d.GetSupplierProcurementRecapsBetweenTime(ctx, from, to)
	if err != nil {
		t.Fatalf("get supplier recaps: %v", err)
	}
	if recaps[0].Total != 300 {
		t.Errorf("revived procurement total = %v, want 300", recaps[0].Total)
	}

	last, err := d.GetLastDrugProcurements(ctx, "D1", 10)
	if err != nil {
		t.Fatalf("get last drug procurements: %v", err)
	}
	if len(last) != 1 {
		t.Errorf("GetLastDrugProcurements returned %d rows, want 1", len(last))
	}
}
