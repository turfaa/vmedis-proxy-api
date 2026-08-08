package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

// TestSoftDeleteProcurementsMissingFromVmedis checks that reconciliation
// soft-deletes exactly the procurements that are in the DB but no longer in
// Vmedis for the reconciled date: procurements still in Vmedis and
// procurements with an invoice date on other dates are left alone.
func TestSoftDeleteProcurementsMissingFromVmedis(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	service := NewService(db, nil, nil, nil, nil)

	date := time.Date(2026, 8, 7, 0, 0, 0, 0, time.Local)
	otherDate := date.AddDate(0, 0, -1)

	newProcurement := func(invoiceNumber string, at time.Time) vmedisv1.Procurement {
		return vmedisv1.Procurement{
			Date:          vmedisv1.Date{Time: at},
			InputTime:     vmedisv1.Time{Time: at},
			InvoiceNumber: invoiceNumber,
			Supplier:      "Supplier A",
			Total:         10,
			ProcurementUnits: []vmedisv1.ProcurementUnit{
				{IDInProcurement: 1, DrugCode: "D1", DrugName: "Drug One", Amount: 5, Unit: "box", Total: 10},
			},
		}
	}

	dumped := []vmedisv1.Procurement{
		newProcurement("OBT1", date),
		newProcurement("OBT2", date),
		newProcurement("OBT-OTHER-DATE", otherDate),
	}
	if err := service.db.UpsertVmedisProcurements(ctx, dumped); err != nil {
		t.Fatalf("dump procurements: %v", err)
	}

	// OBT1 was deleted in Vmedis; OBT2 is still there.
	stillInVmedis := []vmedisv1.Procurement{
		newProcurement("OBT2", date),
	}

	deleted, err := service.softDeleteProcurementsMissingFromVmedis(ctx, date, stillInVmedis)
	if err != nil {
		t.Fatalf("soft delete procurements missing from vmedis: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("soft-deleted %d procurements, want 1", deleted)
	}

	visible, err := service.db.GetProcurementInvoiceNumbersBetweenTime(ctx, date, date.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("get visible invoice numbers: %v", err)
	}
	if len(visible) != 1 || visible[0] != "OBT2" {
		t.Errorf("visible procurements on reconciled date = %v, want [OBT2]", visible)
	}

	otherDateVisible, err := service.db.GetProcurementInvoiceNumbersBetweenTime(ctx, otherDate, otherDate.Add(23*time.Hour))
	if err != nil {
		t.Fatalf("get other date invoice numbers: %v", err)
	}
	if len(otherDateVisible) != 1 || otherDateVisible[0] != "OBT-OTHER-DATE" {
		t.Errorf("procurements on other date = %v, want [OBT-OTHER-DATE]", otherDateVisible)
	}
}
