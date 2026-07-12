package vmedisv1

import (
	"os"
	"reflect"
	"testing"
	"time"
)

// TestParseProcurements parses a trimmed copy of the real
// /laporan-transaksi-pembelian-obat-batch/index page (July 2026 layout, which
// added the "Ketentuan Retur" and "Maks bln sblm ED" detail columns and uses
// the w6 grid widget id).
func TestParseProcurements(t *testing.T) {
	f, err := os.Open("testdata/procurements.html")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	res, err := ParseProcurements(f)
	if err != nil {
		t.Fatalf("ParseProcurements: %v", err)
	}

	if len(res.Procurements) != 2 {
		t.Fatalf("got %d procurements, want 2", len(res.Procurements))
	}

	wantPages := []int{1, 2, 3, 10}
	if !reflect.DeepEqual(res.OtherPages, wantPages) {
		t.Errorf("got other pages %v, want %v", res.OtherPages, wantPages)
	}

	date := func(day, month, year int) Date {
		return Date{Time: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)}
	}

	first := res.Procurements[0]
	wantFirst := Procurement{
		Date:                   date(2, 7, 2026),
		InputTime:              Time{Time: time.Date(2026, 7, 2, 13, 37, 9, 0, time.Local)},
		InvoiceNumber:          "020726",
		Supplier:               "CIMORY",
		Warehouse:              "GUDANG UTAMA",
		PaymentType:            "TUNAI",
		PaymentAccount:         "Kas Umum",
		Operator:               "dini",
		CashDiscountPercentage: Percentage{Value: 0},
		DiscountPercentage:     Percentage{Value: 0},
		DiscountAmount:         0,
		TaxPercentage:          Percentage{Value: 0},
		TaxAmount:              0,
		MiscellaneousCost:      0,
		Total:                  408000,
		ProcurementUnits: []ProcurementUnit{
			{
				IDInProcurement:         1,
				DrugCode:                "OBT2205060007",
				DrugName:                "CIMORY SQUEEZE",
				Amount:                  48,
				Unit:                    "Pcs",
				UnitBasePrice:           8500,
				DiscountPercentage:      Percentage{Value: 0},
				DiscountTwoPercentage:   Percentage{Value: 0},
				DiscountThreePercentage: Percentage{Value: 0},
				TotalUnitPrice:          8500,
				UnitTaxedPrice:          8500,
				ExpiryDate:              date(1, 9, 2029),
				BatchNumber:             "SJSJSIUS",
				Total:                   408000,
			},
		},
	}
	if !reflect.DeepEqual(first, wantFirst) {
		t.Errorf("first procurement mismatch:\ngot  %+v\nwant %+v", first, wantFirst)
	}

	second := res.Procurements[1]
	if second.InvoiceNumber != "604767" || second.Supplier != "SALES YAKULT MOBILAN" {
		t.Errorf("unexpected second procurement header: %+v", second)
	}
	if second.MiscellaneousCost != 300 || second.Total != 278000 {
		t.Errorf("got miscellaneous cost %v and total %v, want 300 and 278000", second.MiscellaneousCost, second.Total)
	}
	if len(second.ProcurementUnits) != 4 {
		t.Fatalf("got %d units in second procurement, want 4", len(second.ProcurementUnits))
	}
	last := second.ProcurementUnits[3]
	if last.DrugName != "YAKULT LIGHT" || last.BatchNumber != "STRO" || last.Total != 35700 {
		t.Errorf("unexpected last unit of second procurement: %+v", last)
	}
}
