package vmedisv1

import (
	"os"
	"reflect"
	"testing"
	"time"
)

// TestParseSales parses a trimmed copy of the real
// /apt-lap-penjualanobat-batch/index page (July 2026 layout, which added the
// "No Order", "Kode Approval", "Promo Penjualan", and "Nama Promo" columns,
// shifting the patient/doctor/salesman/payment columns and pushing "Total"
// from 25 to 31).
func TestParseSales(t *testing.T) {
	f, err := os.Open("testdata/sales.html")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	res, err := ParseSales(f)
	if err != nil {
		t.Fatalf("ParseSales: %v", err)
	}

	if len(res.Sales) != 2 {
		t.Fatalf("got %d sales, want 2", len(res.Sales))
	}

	wantPages := []int{1, 2, 3, 10}
	if !reflect.DeepEqual(res.OtherPages, wantPages) {
		t.Errorf("got other pages %v, want %v", res.OtherPages, wantPages)
	}

	first := res.Sales[0]
	wantFirst := Sale{
		ID:            571870,
		Date:          Time{Time: time.Date(2026, 7, 12, 6, 15, 56, 0, time.Local)},
		Cashier:       "yuli",
		InvoiceNumber: "PJ2607120001",
		PatientName:   "",
		Doctor:        "",
		Salesman:      "Yuli Yuliani",
		Payment:       "TUNAI",
		Total:         6000,
		SaleUnits: []SaleUnit{
			{
				IDInSale:      1,
				DrugCode:      "OBT2605250001",
				DrugName:      "IBUPROFEN 400MG TAB 10x10 (CASPER)",
				Batch:         "F103 (ED : 2028-06-01)",
				Amount:        2,
				Unit:          "Strip",
				UnitPrice:     3000,
				PriceCategory: "Harga Jual 1",
				Discount:      0,
				Tuslah:        0,
				Embalase:      0,
				Total:         6000,
			},
		},
	}
	if !reflect.DeepEqual(first, wantFirst) {
		t.Errorf("first sale mismatch:\ngot  %+v\nwant %+v", first, wantFirst)
	}

	second := res.Sales[1]
	if second.ID != 571873 || second.InvoiceNumber != "PJ2607120004" || second.Salesman != "Rodiansyah" {
		t.Errorf("unexpected second sale header: %+v", second)
	}
	if second.Payment != "TUNAI" || second.Total != 22000 {
		t.Errorf("got payment %q and total %v, want TUNAI and 22000", second.Payment, second.Total)
	}
	if len(second.SaleUnits) != 3 {
		t.Fatalf("got %d units in second sale, want 3", len(second.SaleUnits))
	}
	last := second.SaleUnits[2]
	if last.DrugName != "PARACETAMOL 500MG TAB 10x10 (TRIFA)" || last.Amount != 2 || last.Total != 5000 {
		t.Errorf("unexpected last unit of second sale: %+v", last)
	}
}
