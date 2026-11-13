package report

import (
	"os"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestGenerateDrugQuantitySheet checks that the styling leaves the data as
// it was: the header in the first row, then one row per drug, with the
// quantities as numbers.
func TestGenerateDrugQuantitySheet(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	drugs := []DrugQuantity{
		{DrugName: "ANTASIDA DOEN TAB 10x10 (MUTIFA)", Quantity: 1250, Unit: "Strip"},
		{DrugName: "OBH COMBI ANAK 60ML", Quantity: 37.5, Unit: "Botol"},
		{DrugName: "PARACETAMOL 500MG TAB 10x10 (TRIFA)", Quantity: 0.25, Unit: "Box"},
	}

	if err := GenerateDrugQuantitySheet(f, "Penjualan", drugs); err != nil {
		t.Fatalf("GenerateDrugQuantitySheet: %v", err)
	}

	rows, err := f.GetRows("Penjualan", excelize.Options{RawCellValue: true})
	if err != nil {
		t.Fatalf("get rows: %v", err)
	}

	want := [][]string{
		{"Nama Obat", "Kuantitas", "Satuan"},
		{"ANTASIDA DOEN TAB 10x10 (MUTIFA)", "1250.00", "Strip"},
		{"OBH COMBI ANAK 60ML", "37.50", "Botol"},
		{"PARACETAMOL 500MG TAB 10x10 (TRIFA)", "0.25", "Box"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("got rows\n%q\nwant\n%q", rows, want)
	}

	if path := os.Getenv("REPORT_SAMPLE_PATH"); path != "" {
		if err := f.DeleteSheet("Sheet1"); err != nil {
			t.Fatalf("delete default sheet: %v", err)
		}
		if err := f.SaveAs(path); err != nil {
			t.Fatalf("save sample: %v", err)
		}
	}
}
