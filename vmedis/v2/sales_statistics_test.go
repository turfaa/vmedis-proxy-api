package vmedisv2

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// fakeGateway records the last request and responds with response.
type fakeGateway struct {
	req      GatewayRequest
	response string
}

func (g *fakeGateway) Call(_ context.Context, req GatewayRequest, responseTarget any) error {
	g.req = req
	return json.Unmarshal([]byte(g.response), responseTarget)
}

func TestGetSalesStatistics(t *testing.T) {
	gw := &fakeGateway{response: `{"total":315,"totalall":7138500.5,"grandbiaya":0}`}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	from := time.Date(2026, 5, 20, 13, 0, 0, 0, time.Local)
	to := time.Date(2026, 5, 21, 1, 0, 0, 0, time.Local)

	got, err := c.GetSalesStatistics(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetSalesStatistics: %v", err)
	}

	want := SalesStatistics{NumberOfSales: 315, TotalSales: 7138500.5}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if gw.req.TargetUrl != "apt-lap-penjualanobat-batch/indexcount-v2" || gw.req.TargetVersion != "apivi" || gw.req.TargetOptions.Method != "POST" {
		t.Errorf("got request %+v", gw.req)
	}

	payload := gw.req.Payload.(map[string]any)
	wantPayload := map[string]any{
		"a":            "v1234",
		"login_a":      "v1234",
		"reg":          "ab12",
		"tanggalawal":  "2026-05-20",
		"tanggalakhir": "2026-05-21",
	}
	for k, v := range wantPayload {
		if payload[k] != v {
			t.Errorf("payload[%q] = %v, want %v", k, payload[k], v)
		}
	}
}

func TestGetSalesStatistics_RejectedRequest(t *testing.T) {
	gw := &fakeGateway{response: `{"success":false,"message":"appId is required"}`}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	_, err := c.GetSalesStatistics(context.Background(), time.Now(), time.Now())
	if err == nil || !strings.Contains(err.Error(), "appId is required") {
		t.Fatalf("got error %v, want it to contain the response message", err)
	}
}

func TestGetSalesStatistics_NotConfigured(t *testing.T) {
	gw := &fakeGateway{response: `{"total":1,"totalall":1}`}
	c := &Client{gateway: gw}

	if _, err := c.GetSalesStatistics(context.Background(), time.Now(), time.Now()); err == nil {
		t.Fatal("got no error, want one for missing app id and reg")
	}
}
