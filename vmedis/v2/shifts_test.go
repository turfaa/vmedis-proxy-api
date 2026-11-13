package vmedisv2

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// testShiftsResponse is a trimmed copy of a real response, plus a shift that
// is still open.
const testShiftsResponse = `{"status":1,"total":2,"data":[
	{"shfid":3935,"shfbuka":"2026-11-10 06:15:08","shftutup":"2026-11-10 13:51:55","shfawal":"2376500.00","shfakhir":"5687000.00","shfnote":"UANG LEBIH 3.500. ","uid":57602,"shfuid":28359,"shfuser":null,"shfkode":"20261110-PAGI-rodiansyah","shfpenjualan":"3329500.00","shfreturpenjualan":"0.00","shfpembayaranpiutang":"0.00","shfjurnalmasuk":"0.00","shfjurnalkeluar":"22500.00","shfpembayaranklinik":"0.00","shfpiutangklinik":"0.00","shfreturklinik":"0.00","username":"rodiansyah","diserahkan":"turfaa"},
	{"shfid":3940,"shfbuka":"2026-11-13 06:15:00","shftutup":null,"shfawal":"1000000.00","shfakhir":"0.00","shfnote":null,"shfkode":"20261113-PAGI-andi","shfpenjualan":"500000.00","shfreturpenjualan":"10000.00","shfpembayaranpiutang":"20000.00","shfjurnalmasuk":"30000.00","shfjurnalkeluar":"40000.00","shfpembayaranklinik":"50000.00","shfpiutangklinik":"60000.00","shfreturklinik":"70000.00","username":"andi","diserahkan":null}
]}`

func TestGetAllShiftsBetweenTimes(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"aplaporangantishift/index": func(map[string]any) string { return testShiftsResponse },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	from := time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 11, 13, 23, 59, 59, 0, time.Local)
	got, err := c.GetAllShiftsBetweenTimes(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetAllShiftsBetweenTimes: %v", err)
	}

	want := []Shift{
		{
			ID:                  3935,
			Code:                "20261110-PAGI-rodiansyah",
			Cashier:             "rodiansyah",
			StartedAt:           time.Date(2026, 11, 10, 6, 15, 8, 0, time.Local),
			EndedAt:             time.Date(2026, 11, 10, 13, 51, 55, 0, time.Local),
			InitialCash:         2376500,
			ExpectedFinalCash:   5683500,
			ActualFinalCash:     5687000,
			FinalCashDifference: 3500,
			Supervisor:          "turfaa",
			Notes:               "UANG LEBIH 3.500.",
		},
		{
			ID:        3940,
			Code:      "20261113-PAGI-andi",
			Cashier:   "andi",
			StartedAt: time.Date(2026, 11, 13, 6, 15, 0, 0, time.Local),
			// 1000000 + (500000 + 20000 + 50000 + 60000 + 30000) - (10000 + 70000 + 40000).
			InitialCash:         1000000,
			ExpectedFinalCash:   1540000,
			FinalCashDifference: -1540000,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	payload := gw.calls["aplaporangantishift/index"][0]
	if payload["cari"] != 4 || payload["tglAwal"] != "10 Nov 2026 00:00" || payload["tglAkhir"] != "13 Nov 2026 23:59" {
		t.Errorf("got payload %v", payload)
	}
	if payload["offset"] != 0 || payload["limit"] != shiftsPageSize {
		t.Errorf("got payload %v", payload)
	}
}

func TestGetAllShiftsBetweenTimesPaginates(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"aplaporangantishift/index": func(p map[string]any) string {
			n := shiftsPageSize
			if p["offset"] != 0 {
				n = 1
			}
			rows := make([]map[string]any, n)
			for i := range rows {
				rows[i] = map[string]any{"shfid": p["offset"].(int) + i, "shfbuka": "2026-11-11 06:00:00"}
			}
			b, _ := json.Marshal(map[string]any{"status": 1, "data": rows})
			return string(b)
		},
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllShiftsBetweenTimes(context.Background(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("GetAllShiftsBetweenTimes: %v", err)
	}

	if len(got) != shiftsPageSize+1 {
		t.Errorf("got %d shifts, want %d", len(got), shiftsPageSize+1)
	}
	if calls := gw.calls["aplaporangantishift/index"]; len(calls) != 2 || calls[1]["offset"] != shiftsPageSize {
		t.Errorf("got calls %v", calls)
	}
}
