package vmedisv2

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

// routingGateway responds to each target URL with the response returned by
// its handler, and records the payloads it was called with.
type routingGateway struct {
	mu       sync.Mutex
	handlers map[string]func(payload map[string]any) string
	calls    map[string][]map[string]any
}

func (g *routingGateway) Call(_ context.Context, req GatewayRequest, responseTarget any) error {
	payload := req.Payload.(map[string]any)

	g.mu.Lock()
	if g.calls == nil {
		g.calls = map[string][]map[string]any{}
	}
	g.calls[req.TargetUrl] = append(g.calls[req.TargetUrl], payload)
	handler, ok := g.handlers[req.TargetUrl]
	g.mu.Unlock()

	if !ok {
		return fmt.Errorf("unexpected target %s", req.TargetUrl)
	}

	return json.Unmarshal([]byte(handler(payload)), responseTarget)
}

// The responses below are trimmed copies of real responses.
const (
	testSalesmenResponse = `[{"nama_lengkap":"Yola Alianingsih","user_id":"57603"}]`

	testSalesResponse = `[
		{"pjid":638892,"pjtanggal":"2026-11-11T06:22:28.000Z","username":"rodiansyah","pjnofaktur":"PJ2611110001","pasnama":null,"doknama":null,"pjsales":"57603","pjjenis":"tunai","grandtotal":29000,"total":29000},
		{"pjid":638893,"pjtanggal":"2026-11-11T18:05:52.000Z","username":"naisila","pjnofaktur":"PJ2611110002","pasnama":"Budi ","doknama":"dr. Ani ","pjsales":"","pjjenis":"platform","grandtotal":"1500.50","total":1500.5}
	]`

	testSaleUnitsResponse = `[
		{"pjdid":1319638,"obatkode":"OBT2","pjdnama":"TOLAK LINU ","pjdnobatch":"[\" 0489 (ED : 2028-06-01)\",\" 0490 (ED : 2028-07-01)\"]","pjdjumlah":"1.00","sonama":"Strip","pjdhargajual":"18500.00","pjdhrgid":"120005","pjdhrgcombo":1,"pjddiskon":"4.3243243243","pjddiskontext":"800.00","pjdtuslah":"0","pjdemba":"0","pjdtotal":"17700.00"},
		{"pjdid":1319637,"obatkode":"","pjdnama":"RACIKAN","pjdnobatch":"","pjdjumlah":"0.00","sonama":"Strip","pjdhargajual":"0.00","pjdhrgid":"resep1","pjdhrgcombo":1,"pjddiskontext":"0.00","pjdtuslah":"0","pjdemba":"0","pjdtotal":"0.00"},
		{"pjdid":1319639,"obatkode":"OBT3","pjdnama":"COMBIVENT","pjdnobatch":"[]","pjdjumlah":"2.00","sonama":"Ampul","pjdhargajual":"7000.00","pjdhrgid":"121895_resep1","pjdhrgcombo":3,"pjddiskontext":"0.00","pjdtuslah":"2000","pjdemba":"500","pjdtotal":"16500.00"}
	]`
)

func TestGetAllSalesBetweenDates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"appenjualanobat-batch/datasales":      func(map[string]any) string { return testSalesmenResponse },
		"apt-lap-penjualanobat-batch/index-v3": func(map[string]any) string { return testSalesResponse },
		"apt-lap-penjualanobat-batch/indexdetail-v2": func(p map[string]any) string {
			if p["pjid"] == 638892 {
				return testSaleUnitsResponse
			}
			return `[]`
		},
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	day := time.Date(2026, 11, 11, 0, 0, 0, 0, time.Local)
	got, err := c.GetAllSalesBetweenDates(context.Background(), day, day)
	if err != nil {
		t.Fatalf("GetAllSalesBetweenDates: %v", err)
	}

	want := []Sale{
		{
			ID:            638892,
			Date:          time.Date(2026, 11, 11, 6, 22, 28, 0, time.Local),
			Cashier:       "rodiansyah",
			InvoiceNumber: "PJ2611110001",
			Salesman:      "Yola Alianingsih",
			Payment:       "TUNAI",
			Total:         29000,
			SaleUnits: []SaleUnit{
				{IDInSale: 1, DrugName: "RACIKAN", Unit: "Strip"},
				{IDInSale: 2, DrugCode: "OBT2", DrugName: "TOLAK LINU", Batch: "0489 (ED : 2028-06-01), 0490 (ED : 2028-07-01)", Amount: 1, Unit: "Strip", UnitPrice: 18500, PriceCategory: "Harga Jual 1", Discount: 800, Total: 17700},
				{IDInSale: 3, DrugCode: "OBT3", DrugName: "COMBIVENT", Amount: 2, Unit: "Ampul", UnitPrice: 7000, PriceCategory: "Harga Jual 3", Tuslah: 2000, Embalase: 500, Total: 16500},
			},
		},
		{
			ID:            638893,
			Date:          time.Date(2026, 11, 11, 18, 5, 52, 0, time.Local),
			Cashier:       "naisila",
			InvoiceNumber: "PJ2611110002",
			PatientName:   "Budi",
			Doctor:        "dr. Ani",
			Payment:       "PLATFORM ONLINE",
			Total:         1500.5,
			SaleUnits:     []SaleUnit{},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	page := gw.calls["apt-lap-penjualanobat-batch/index-v3"][0]
	if page["tanggalawal"] != "2026-11-11" || page["tanggalakhir"] != "2026-11-11" || page["offset"] != 0 || page["a"] != "v1234" || page["reg"] != "ab12" {
		t.Errorf("got sales page payload %v", page)
	}
}

func TestGetAllSalesBetweenDates_Paginates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"appenjualanobat-batch/datasales": func(map[string]any) string { return `[]` },
		"apt-lap-penjualanobat-batch/index-v3": func(p map[string]any) string {
			// A full first page, then a partial one.
			n := salesPageSize
			if p["offset"] != 0 {
				n = 1
			}
			rows := make([]map[string]any, n)
			for i := range rows {
				rows[i] = map[string]any{"pjid": p["offset"].(int) + i, "pjtanggal": "2026-11-11T06:22:28.000Z"}
			}
			b, _ := json.Marshal(rows)
			return string(b)
		},
		"apt-lap-penjualanobat-batch/indexdetail-v2": func(map[string]any) string { return `[]` },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllSalesBetweenDates(context.Background(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("GetAllSalesBetweenDates: %v", err)
	}

	if len(got) != salesPageSize+1 {
		t.Errorf("got %d sales, want %d", len(got), salesPageSize+1)
	}
	if pages := len(gw.calls["apt-lap-penjualanobat-batch/index-v3"]); pages != 2 {
		t.Errorf("got %d page requests, want 2", pages)
	}
}

func TestNumber_UnmarshalJSON(t *testing.T) {
	for in, want := range map[string]number{`6500`: 6500, `"6500.00"`: 6500, `"4.5"`: 4.5, `null`: 0, `""`: 0} {
		var got number
		if err := json.Unmarshal([]byte(in), &got); err != nil {
			t.Errorf("unmarshal %s: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("unmarshal %s: got %v, want %v", in, got, want)
		}
	}
}
