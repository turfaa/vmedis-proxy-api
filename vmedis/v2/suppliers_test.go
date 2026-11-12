package vmedisv2

import (
	"context"
	"reflect"
	"testing"
)

// testSuppliersResponse is a trimmed copy of a real response.
const testSuppliersResponse = `[
	{"supid":231,"supkode":"SUP2403190229","supnama":"ADIWARNA TUNGGAL JAYA ","supalamat":"Jl. Contoh 1","supkota":"Bandung","suptelp":"022-123","suphp":"0812","supnorek":"123-456","supnpwp":"01.234","supaktif":1,"supleadtime":3,"supreting":"4.50","level_no":3},
	{"supid":15,"supkode":"SUP2205050014","supnama":"<GAK TAU>","supalamat":null,"supkota":"","suptelp":"","suphp":"","supnorek":"","supnpwp":"","supaktif":0,"supleadtime":0,"supreting":"0.00","level_no":1}
]`

func TestGetAllSuppliers(t *testing.T) {
	gw := &routingGateway{handlers: map[string]func(map[string]any) string{
		"data-supplier/list": func(map[string]any) string { return testSuppliersResponse },
	}}
	c := &Client{gateway: gw, appID: "v1234", reg: "ab12"}

	got, err := c.GetAllSuppliers(context.Background())
	if err != nil {
		t.Fatalf("GetAllSuppliers: %v", err)
	}

	want := []Supplier{
		{ID: 231, Code: "SUP2403190229", Name: "ADIWARNA TUNGGAL JAYA", Address: "Jl. Contoh 1", City: "Bandung", Phone: "022-123", MobilePhone: "0812", BankAccountNumber: "123-456", TaxNumber: "01.234", Active: true, LeadTimeDays: 3, Rating: 4.5},
		{ID: 15, Code: "SUP2205050014", Name: "<GAK TAU>"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	payload := gw.calls["data-supplier/list"][0]
	if _, ok := payload["supaktif"]; ok {
		t.Errorf("got payload %v, want no supaktif filter so inactive suppliers are listed too", payload)
	}
	if payload["offset"] != 0 || payload["limit"] != suppliersPageSize {
		t.Errorf("got payload %v", payload)
	}
}
