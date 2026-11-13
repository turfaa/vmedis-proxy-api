package vmedisv2

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// testStockOpnamesResponse is a trimmed copy of a real response, plus a row
// with the nullable fields null.
const testStockOpnamesResponse = `{"status":1,"data":[
	{"sopid":99310,"soptanggal":"2026-10-02T09:13:36.000Z","sopkode":"SOP202610020002","obatkode":"A00349","sodbarcode":"8993034780074","obatnama":"POLYCROL FORTE GEL 180ML","sonama":"Botol","gudnama":"GUDANG UTAMA","katonama":"-","loknama":"LEMARI A3","golonama":"DEPAN","obatjenis":"NON KONSINYASI","sopselisih":"-2.00","sopjumlah":"0.00","stok":"2.00","sophrgpp":"35853.00","sophrgbeli":"34000.00","sophrgjual1":"40500.00","sophrgdiskon1":"0.00","sophrgjual2":"39500.00","sopketerangan":"ed oktober 2026","sopjmled":"0.00","sophrgdiskon2":"0.00","sophrgjual3":"41231.00","sophrgdiskon3":"0.00","sopnobatch":"1348","soptglexpired":"2026-10-01T00:00:00.000Z","username":"dini"},
	{"sopid":99400,"soptanggal":"2026-10-03T20:05:07.000Z","sopkode":"SOP202610030001 ","obatkode":"OBT1","obatnama":"DRUG ","sonama":"Strip","gudnama":null,"sopselisih":1,"sopjumlah":4,"stok":3,"sophrgpp":"1000.50","sophrgjual1":"1500.00","sopketerangan":null,"sopnobatch":null,"soptglexpired":null,"username":null}
]}`

func TestGetAllStockOpnamesBetweenDates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-lap-stok-opname-batch/index": func(map[string]any) string { return testStockOpnamesResponse },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	from := time.Date(2026, 10, 2, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 10, 3, 23, 59, 59, 0, time.Local)
	got, err := c.GetAllStockOpnamesBetweenDates(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetAllStockOpnamesBetweenDates: %v", err)
	}

	want := []StockOpname{
		{
			ID:                  99310,
			Code:                "SOP202610020002",
			Time:                time.Date(2026, 10, 2, 9, 13, 36, 0, time.Local),
			DrugCode:            "A00349",
			DrugName:            "POLYCROL FORTE GEL 180ML",
			Unit:                "Botol",
			Warehouse:           "GUDANG UTAMA",
			BatchCode:           "1348",
			ExpiryDate:          time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local),
			InitialQuantity:     2,
			RealQuantity:        0,
			QuantityDifference:  -2,
			HPP:                 35853,
			SalePrice:           40500,
			HPPDifference:       -71706,
			SalePriceDifference: -81000,
			Notes:               "ed oktober 2026",
			Operator:            "dini",
		},
		{
			ID:                  99400,
			Code:                "SOP202610030001",
			Time:                time.Date(2026, 10, 3, 20, 5, 7, 0, time.Local),
			DrugCode:            "OBT1",
			DrugName:            "DRUG",
			Unit:                "Strip",
			InitialQuantity:     3,
			RealQuantity:        4,
			QuantityDifference:  1,
			HPP:                 1000.5,
			SalePrice:           1500,
			HPPDifference:       1000.5,
			SalePriceDifference: 1500,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	payload := gw.calls["apt-lap-stok-opname-batch/index"][0]
	if payload["cari"] != 4 || payload["tglAwal"] != "02 Oct 2026" || payload["tglAkhir"] != "03 Oct 2026" {
		t.Errorf("got payload %v", payload)
	}
	if payload["offset"] != 0 || payload["limit"] != stockOpnamesPageSize || payload["sorting"] != "sopid ASC" {
		t.Errorf("got payload %v", payload)
	}
}

func TestGetAllStockOpnamesBetweenDatesPaginates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-lap-stok-opname-batch/index": func(p map[string]any) string {
			n := stockOpnamesPageSize
			if p["offset"] != 0 {
				n = 1
			}
			rows := make([]map[string]any, n)
			for i := range rows {
				rows[i] = map[string]any{"sopid": p["offset"].(int) + i, "soptanggal": "2026-10-02T09:13:36.000Z"}
			}
			b, _ := json.Marshal(map[string]any{"status": 1, "data": rows})
			return string(b)
		},
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllStockOpnamesBetweenDates(context.Background(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("GetAllStockOpnamesBetweenDates: %v", err)
	}

	if len(got) != stockOpnamesPageSize+1 {
		t.Errorf("got %d stock opnames, want %d", len(got), stockOpnamesPageSize+1)
	}
	if calls := gw.calls["apt-lap-stok-opname-batch/index"]; len(calls) != 2 || calls[1]["offset"] != stockOpnamesPageSize {
		t.Errorf("got calls %v", calls)
	}
}

func TestGetAllStockOpnamesBetweenDatesRejectsBadTime(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"apt-lap-stok-opname-batch/index": func(map[string]any) string {
			return `{"status":1,"data":[{"sopid":1,"sopkode":"SOP1","soptanggal":"02/10/2026"}]}`
		},
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	if _, err := c.GetAllStockOpnamesBetweenDates(context.Background(), time.Now(), time.Now()); err == nil {
		t.Error("got no error, want an error for an unparseable time")
	}
}
