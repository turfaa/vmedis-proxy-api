package sale

import (
	"context"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	vmedisv1 "github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

type nopDrugProducer struct{}

func (nopDrugProducer) ProduceUpdatedDrugByVmedisCode(context.Context, []*kafkapb.UpdatedDrugByVmedisCode) error {
	return nil
}

// TestSoftDeleteSalesMissingFromVmedis checks that reconciliation soft-deletes
// exactly the sales that are in the DB but no longer in Vmedis for the
// reconciled date: sales still in Vmedis and sales sold on other dates are
// left alone, and duplicated invoice numbers are matched through the same
// de-duplication the dump applies.
func TestSoftDeleteSalesMissingFromVmedis(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	service := NewService(db, nil, nil, nopDrugProducer{})

	date := time.Date(2026, 8, 7, 0, 0, 0, 0, time.Local)
	soldAt := date.Add(10 * time.Hour)
	otherDateSoldAt := date.AddDate(0, 0, -1).Add(10 * time.Hour)

	newSale := func(id int, invoiceNumber string, at time.Time) vmedisv1.Sale {
		return vmedisv1.Sale{
			ID:            id,
			Date:          vmedisv1.Time{Time: at},
			InvoiceNumber: invoiceNumber,
			Total:         10,
			SaleUnits: []vmedisv1.SaleUnit{
				{IDInSale: 1, DrugCode: "D1", DrugName: "Drug One", Amount: 1, Unit: "tablet", Total: 10},
			},
		}
	}

	// PJ2 and PJ2 share an invoice number, so the dump stores them as PJ2 and PJ2-2.
	dumped := []vmedisv1.Sale{
		newSale(1, "PJ1", soldAt),
		newSale(2, "PJ2", soldAt),
		newSale(3, "PJ2", soldAt),
		newSale(4, "PJ-OTHER-DATE", otherDateSoldAt),
	}
	if err := service.dumpSalesToDB(ctx, dumped); err != nil {
		t.Fatalf("dump sales: %v", err)
	}

	// PJ1 was deleted in Vmedis; the duplicated PJ2 pair is still there.
	stillInVmedis := []vmedisv1.Sale{
		newSale(2, "PJ2", soldAt),
		newSale(3, "PJ2", soldAt),
	}

	deleted, err := service.softDeleteSalesMissingFromVmedis(ctx, date, stillInVmedis)
	if err != nil {
		t.Fatalf("soft delete sales missing from vmedis: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("soft-deleted %d sales, want 1", deleted)
	}

	visible, err := service.db.GetSaleInvoiceNumbersBetweenTime(ctx, date, date.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("get visible invoice numbers: %v", err)
	}
	want := map[string]bool{"PJ2": true, "PJ2-2": true}
	if len(visible) != len(want) {
		t.Fatalf("visible sales on reconciled date = %v, want %v", visible, want)
	}
	for _, invoiceNumber := range visible {
		if !want[invoiceNumber] {
			t.Errorf("sale %s is visible, want only %v", invoiceNumber, want)
		}
	}

	otherDate, err := service.db.GetSaleInvoiceNumbersBetweenTime(ctx, otherDateSoldAt.Add(-time.Hour), otherDateSoldAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("get other date invoice numbers: %v", err)
	}
	if len(otherDate) != 1 || otherDate[0] != "PJ-OTHER-DATE" {
		t.Errorf("sales on other date = %v, want [PJ-OTHER-DATE]", otherDate)
	}
}
