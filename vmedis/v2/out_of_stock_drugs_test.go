package vmedisv2

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

// testOutOfStockDrugsResponse is a trimmed copy of a real response.
const testOutOfStockDrugsResponse = `{"status":1,"data":[
	{"obatid":1078,"obatkode":"OBT2205090002","obatnama":"ACARBOSE 50MG (DEXA)","obatminstok":"7.00","gudnama":"GUDANG UTAMA","stok":"1.00","pabnama":"DEXAMEDICA ","sonama":"Strip","supnama":"PT. CARMELLA GUSTAVIAN SETIABUDI"},
	{"obatid":9,"obatkode":"MAN202204230052","obatnama":"OLD DRUG","obatminstok":"0.00","gudnama":"GUDANG UTAMA","stok":"-2.00","pabnama":null,"sonama":"(Jangan Dipakai)","supnama":null}
]}`

func TestGetAllOutOfStockDrugs(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"lap-obat-habis/list": func(map[string]any) string { return testOutOfStockDrugsResponse },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllOutOfStockDrugs(context.Background())
	if err != nil {
		t.Fatalf("GetAllOutOfStockDrugs: %v", err)
	}

	want := []OutOfStockDrug{
		{
			VmedisCode:   "OBT2205090002",
			Name:         "ACARBOSE 50MG (DEXA)",
			Manufacturer: "DEXAMEDICA",
			Supplier:     "PT. CARMELLA GUSTAVIAN SETIABUDI",
			MinimumStock: Stock{Unit: "Strip", Quantity: 7},
			Stock:        Stock{Unit: "Strip", Quantity: 1},
		},
		{
			VmedisCode:   "MAN202204230052",
			Name:         "OLD DRUG",
			MinimumStock: Stock{Unit: "(Jangan Dipakai)"},
			Stock:        Stock{Unit: "(Jangan Dipakai)", Quantity: -2},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	payload := gw.calls["lap-obat-habis/list"][0]
	if payload["gudid"] != "pergudang" || payload["obathabisnotif"] != "1" || payload["obatstatus"] != "1" || payload["offset"] != 0 {
		t.Errorf("got payload %v", payload)
	}
}

func TestGetAllOutOfStockDrugs_Paginates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"lap-obat-habis/list": func(p map[string]any) string {
			n := outOfStockDrugsPageSize
			if p["offset"] != 0 {
				n = 1
			}
			rows := make([]map[string]any, n)
			for i := range rows {
				rows[i] = map[string]any{"obatkode": "OBT", "stok": "0"}
			}
			b, _ := json.Marshal(map[string]any{"status": 1, "data": rows})
			return string(b)
		},
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllOutOfStockDrugs(context.Background())
	if err != nil {
		t.Fatalf("GetAllOutOfStockDrugs: %v", err)
	}

	if len(got) != outOfStockDrugsPageSize+1 {
		t.Errorf("got %d drugs, want %d", len(got), outOfStockDrugsPageSize+1)
	}
	if pages := len(gw.calls["lap-obat-habis/list"]); pages != 2 {
		t.Errorf("got %d page requests, want 2", pages)
	}
}
