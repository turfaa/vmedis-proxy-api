package vmedisv2

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

// testDrugViewResponse is a trimmed copy of a real apt-obat-batch/view
// response, with the units out of order and a second batch added.
const testDrugViewResponse = `{"status":1,"message":"","data":{
	"obat":{"obatid":2281,"obatkode":"OBT2312090001","obatnama":"ANTASIDA DOEN TAB 10x10 (MUTIFA) ","kodekfa":null,"pabnama":"MUTIFA","obatminstok":"511.00"},
	"satuan":[
		{"sodid":3838,"soid":6,"sodkonversi":"10.00","sodhasilkonversi":"10.00","sodurut":2,"sonama":"Box","hrgjual1":"13000.00","hrgjual2":"12500.00","hrgjual3":"14000.00"},
		{"sodid":3837,"soid":2,"sodkonversi":"1.00","sodhasilkonversi":"1.00","sodurut":1,"sonama":"Strip","hrgjual1":"2000.00","hrgjual2":"2000.00","hrgjual3":"3500.00"}
	],
	"gudang_lokasi":{"gudid":1,"gudnama":"GUDANG UTAMA"},
	"cart":[
		{"mshid":1450683,"sodurut":1,"soid":2,"sonama":"Strip","stok":163,"batch":"6046","tglexpired":"2028-02-01"},
		{"mshid":1450684,"sodurut":2,"soid":6,"sonama":"Box","stok":"2","batch":"6047","tglexpired":"2028-03-01"},
		{"mshid":1450685,"sodurut":1,"soid":2,"sonama":"Strip","stok":7,"batch":"6048","tglexpired":"2028-04-01"},
		{"mshid":1450686,"sodurut":2,"soid":6,"sonama":"Box","stok":0,"batch":"6049","tglexpired":"2028-05-01"}
	]
}}`

func TestGetDrug(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-obat-batch/view": func(map[string]any) string { return testDrugViewResponse },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetDrug(context.Background(), 2281)
	if err != nil {
		t.Fatalf("GetDrug: %v", err)
	}

	want := Drug{
		VmedisID:     2281,
		VmedisCode:   "OBT2312090001",
		Name:         "ANTASIDA DOEN TAB 10x10 (MUTIFA)",
		Manufacturer: "MUTIFA",
		MinimumStock: Stock{Unit: "Strip", Quantity: 511},
		Units: []Unit{
			{Unit: "Strip", UnitOrder: 0, PriceOne: 2000, PriceTwo: 2000, PriceThree: 3500},
			{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 1, PriceOne: 13000, PriceTwo: 12500, PriceThree: 14000},
		},
		Stocks: []Stock{
			{Unit: "Box", Quantity: 2},
			{Unit: "Strip", Quantity: 170},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	if payload := gw.calls["apt-obat-batch/view"][0]; payload["obatid"] != int64(2281) || payload["gudid"] != "" {
		t.Errorf("got payload %v", payload)
	}
}

func TestGetDrugRejected(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-obat-batch/view": func(map[string]any) string { return `{"status":0,"message":"Data tidak ditemukan"}` },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	if _, err := c.GetDrug(context.Background(), 1); err == nil {
		t.Error("got no error, want one for status 0")
	}
}

// TestGetAllDrugs checks that all the pages are fetched up to the total, even
// when the endpoint returns fewer drugs per page than asked for, as the real
// one does past 100.
func TestGetAllDrugs(t *testing.T) {
	const (
		total          = 90
		serverPageSize = 40
	)

	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-obat-batch/index": func(p map[string]any) string {
			first := (p["page"].(int) - 1) * serverPageSize
			var rows []map[string]any
			for id := first + 1; id <= min(first+serverPageSize, total); id++ {
				rows = append(rows, map[string]any{"obatid": id, "obatkode": "OBT", "obatnama": "DRUG", "kodekfa": nil, "pabnama": "PAB", "obatminstok": "5.00"})
			}
			b, _ := json.Marshal(map[string]any{"status": 1, "data": rows, "total": total})
			return string(b)
		},
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllDrugs(context.Background())
	if err != nil {
		t.Fatalf("GetAllDrugs: %v", err)
	}

	if len(got) != total {
		t.Fatalf("got %d drugs, want %d", len(got), total)
	}
	if want := (Drug{VmedisID: 1, VmedisCode: "OBT", Name: "DRUG", Manufacturer: "PAB", MinimumStock: Stock{Quantity: 5}}); !reflect.DeepEqual(got[0], want) {
		t.Errorf("got first drug %+v, want %+v", got[0], want)
	}

	calls := gw.calls["apt-obat-batch/index"]
	if len(calls) != 3 || calls[2]["page"] != 3 || calls[0]["obatstatus"] != "1" {
		t.Errorf("got calls %v", calls)
	}
}

func TestGetAllDrugsEmpty(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-obat-batch/index": func(map[string]any) string { return `{"status":1,"data":[],"total":0}` },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	if _, err := c.GetAllDrugs(context.Background()); err == nil {
		t.Error("got no error, want one for an empty drug list")
	}
}
