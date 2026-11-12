package vmedisv2

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// The responses below are trimmed copies of real responses.
const (
	testProcurementsResponse = `{"status":1,"data":[
		{"pemoid":14419,"pemotanggal":"2026-11-11 12:30:08","pemotglinput":"2026-11-12 12:30:08","pemofaktur":" A26-14112 ","supnama":"ADIWARNA TUNGGAL JAYA","gudnama":"GUDANG UTAMA","pemojenis":"HUTANG","aknama":null,"username":"adi","pemodiskontunai":"0.00","pemodiskon":"2.00","pemodiskontext":"9435.00","pemopajak":"11.00","pemopajaktext":"50854.00","pemobiaya":"0.00","pemograndtotal":"513159.00"}
	]}`

	testProcurementUnitsResponse = `{"status":1,"data":[
		{"pemdoid":57160,"obatkode":"OBT2","obatnama":"LARUTAN ","sonama":"KARTON","pemdoexpired":null,"pemdonobatch":"2345","pemdoharga":"126000.00","pemdojumlah":"2.00","pemdodiskon":"0.00","pemdodiskon2":"5.00","pemdodiskon3":"1.00","pemdototal":"252000.00","pemdohpp":"126000.00","pemdohnappn":"139860.00"},
		{"pemdoid":57159,"obatkode":"OBT1","obatnama":"POPOK","sonama":"KARTON","pemdoexpired":"2028-01-01T00:00:00.000Z","pemdonobatch":"8812","pemdoharga":"471740.00","pemdojumlah":"2.00","pemdodiskon":"50.00","pemdodiskon2":"0.00","pemdodiskon3":"0.00","pemdototal":"471740.00","pemdohpp":"256579.39","pemdohnappn":"523631.40"}
	]}`
)

func TestGetAllProcurementsBetweenDates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"laporan-transaksi-pembelian-obat-batch/index-v2": func(map[string]any) string { return testProcurementsResponse },
		"laporan-transaksi-pembelian-obat-batch/detail":   func(map[string]any) string { return testProcurementUnitsResponse },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	day := time.Date(2026, 11, 11, 0, 0, 0, 0, time.Local)
	got, err := c.GetAllProcurementsBetweenDates(context.Background(), day, day)
	if err != nil {
		t.Fatalf("GetAllProcurementsBetweenDates: %v", err)
	}

	want := []Procurement{{
		ID:                 14419,
		Date:               time.Date(2026, 11, 11, 0, 0, 0, 0, time.Local),
		InputTime:          time.Date(2026, 11, 12, 12, 30, 8, 0, time.Local),
		InvoiceNumber:      "A26-14112",
		Supplier:           "ADIWARNA TUNGGAL JAYA",
		Warehouse:          "GUDANG UTAMA",
		PaymentType:        "HUTANG",
		Operator:           "adi",
		DiscountPercentage: 2,
		DiscountAmount:     9435,
		TaxPercentage:      11,
		TaxAmount:          50854,
		Total:              513159,
		ProcurementUnits: []ProcurementUnit{
			{IDInProcurement: 1, DrugCode: "OBT1", DrugName: "POPOK", Amount: 2, Unit: "KARTON", UnitBasePrice: 471740, DiscountPercentage: 50, TotalUnitPrice: 256579.39, UnitTaxedPrice: 523631.4, ExpiryDate: time.Date(2028, 1, 1, 0, 0, 0, 0, time.Local), BatchNumber: "8812", Total: 471740},
			{IDInProcurement: 2, DrugCode: "OBT2", DrugName: "LARUTAN", Amount: 2, Unit: "KARTON", UnitBasePrice: 126000, DiscountTwoPercentage: 5, DiscountThreePercentage: 1, TotalUnitPrice: 126000, UnitTaxedPrice: 139860, BatchNumber: "2345", Total: 252000},
		},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	page := gw.calls["laporan-transaksi-pembelian-obat-batch/index-v2"][0]
	if page["tanggalawal"] != "2026-11-11" || page["tanggalakhir"] != "2026-11-11" || page["offset"] != 0 || page["app_jenis"] != 1 {
		t.Errorf("got procurements page payload %v", page)
	}
	if detail := gw.calls["laporan-transaksi-pembelian-obat-batch/detail"][0]; detail["id"] != 14419 {
		t.Errorf("got detail payload %v", detail)
	}
}

func TestGetAllProcurementsBetweenDates_Paginates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"laporan-transaksi-pembelian-obat-batch/index-v2": func(p map[string]any) string {
			n := procurementsPageSize
			if p["offset"] != 0 {
				n = 1
			}
			rows := make([]map[string]any, n)
			for i := range rows {
				rows[i] = map[string]any{"pemoid": p["offset"].(int) + i, "pemotanggal": "2026-11-11 12:30:08", "pemotglinput": "2026-11-11 12:30:08"}
			}
			b, _ := json.Marshal(map[string]any{"status": 1, "data": rows})
			return string(b)
		},
		"laporan-transaksi-pembelian-obat-batch/detail": func(map[string]any) string { return `{"status":1,"data":[]}` },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllProcurementsBetweenDates(context.Background(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("GetAllProcurementsBetweenDates: %v", err)
	}

	if len(got) != procurementsPageSize+1 {
		t.Errorf("got %d procurements, want %d", len(got), procurementsPageSize+1)
	}
	if pages := len(gw.calls["laporan-transaksi-pembelian-obat-batch/index-v2"]); pages != 2 {
		t.Errorf("got %d page requests, want 2", pages)
	}
}
