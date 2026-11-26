package vmedisv2

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// journalFixture is one decrypted gateway call of the vmedis Jurnal Keuangan
// page, anonymised, from testdata/journal_*.json.
type journalFixture struct {
	TargetUrl     string          `json:"targetUrl"`
	TargetVersion string          `json:"targetVersion"`
	Request       map[string]any  `json:"request"`
	Response      json.RawMessage `json:"response"`
}

func loadJournalFixture(t *testing.T, name string) journalFixture {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var f journalFixture
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("parse fixture %s: %v", name, err)
	}
	return f
}

// newJournalTestClient returns a client with the fixtures' apotek and user,
// answering each fixture's target with its response.
func newJournalTestClient(fixtures ...journalFixture) (*Client, *recordingGateway) {
	gw := &recordingGateway{routingGateway: routingGateway{handlers: map[string]func(map[string]any) string{}}}
	for _, f := range fixtures {
		response := string(f.Response)
		gw.handlers[f.TargetUrl] = func(map[string]any) string { return response }
	}

	return (&Client{gateway: gw, appID: "v0000", reg: "ab12"}).WithUserID(1001), gw
}

// recordingGateway is a routingGateway that also records the requests.
type recordingGateway struct {
	routingGateway
	requests []GatewayRequest
}

func (g *recordingGateway) Call(ctx context.Context, req GatewayRequest, responseTarget any) error {
	g.mu.Lock()
	g.requests = append(g.requests, req)
	g.mu.Unlock()

	return g.routingGateway.Call(ctx, req, responseTarget)
}

// normalise round-trips v through JSON, so payloads built with Go ints compare
// equal to fixtures decoded with float64s.
func normalise(t *testing.T, v any) map[string]any {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func assertJournalRequest(t *testing.T, gw *recordingGateway, f journalFixture, ignore ...string) {
	t.Helper()

	calls := gw.calls[f.TargetUrl]
	if len(calls) != 1 {
		t.Fatalf("got %d calls to %s, want 1", len(calls), f.TargetUrl)
	}

	got, want := normalise(t, calls[0]), normalise(t, f.Request)
	for _, k := range ignore {
		delete(got, k)
		delete(want, k)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s payload\ngot  %v\nwant %v", f.TargetUrl, got, want)
	}

	for _, req := range gw.requests {
		if req.TargetUrl == f.TargetUrl && req.TargetVersion != f.TargetVersion {
			t.Errorf("%s got Target-Version %q, want %q", f.TargetUrl, req.TargetVersion, f.TargetVersion)
		}
	}
}

func TestListJournalAccounts(t *testing.T) {
	f := loadJournalFixture(t, "journal_account_options.json")
	c, gw := newJournalTestClient(f)

	got, err := c.ListJournalAccounts(context.Background())
	if err != nil {
		t.Fatalf("ListJournalAccounts: %v", err)
	}

	want := []JournalAccount{
		{ID: 180, Code: "110001", Name: "Kas Umum"},
		{ID: 618, Code: "111004", Name: "Bank OCBC"},
		{ID: 592, Code: "171001", Name: "Piutang Pegawai"},
		{ID: 377, Code: "610002", Name: "Gaji Karyawan"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	assertJournalRequest(t, gw, f)
}

func TestReserveJournalVoucherNo(t *testing.T) {
	f := loadJournalFixture(t, "journal_get_code.json")
	c, gw := newJournalTestClient(f)

	got, err := c.ReserveJournalVoucherNo(context.Background(), time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("ReserveJournalVoucherNo: %v", err)
	}
	if got != "JRLM2610310001" {
		t.Errorf("got %q, want JRLM2610310001", got)
	}

	assertJournalRequest(t, gw, f)
}

func TestCreateJournalEntry(t *testing.T) {
	f := loadJournalFixture(t, "journal_create.json")
	c, gw := newJournalTestClient(f)

	got, err := c.CreateJournalEntry(context.Background(), NewJournalEntry{
		Date:            time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local),
		VoucherNo:       "JRLM2610310001",
		Description:     "Gaji Karyawan Contoh Oktober 2026 (Snapshot #42)",
		DebitAccountID:  377,
		CreditAccountID: 618,
		Amount:          2500000,
	})
	if err != nil {
		t.Fatalf("CreateJournalEntry: %v", err)
	}

	want := CreatedJournalEntry{ID: 1000001, VoucherNo: "JRLM2610310001"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	// The time is the current time.
	assertJournalRequest(t, gw, f, "time")
	if _, err := time.Parse(time.TimeOnly, gw.calls[f.TargetUrl][0]["time"].(string)); err != nil {
		t.Errorf("time: %v", err)
	}

	if !gw.requests[0].NoRetry {
		t.Error("got a retried create, want it sent once")
	}
}

func TestCreateJournalEntry_Failure(t *testing.T) {
	c, _ := newJournalTestClient(journalFixture{
		TargetUrl: "jurnal/create",
		Response:  json.RawMessage(`{"success":false,"message":"No. Bukti sudah dipakai"}`),
	})

	_, err := c.CreateJournalEntry(context.Background(), NewJournalEntry{Amount: 1, DebitAccountID: 1, CreditAccountID: 2})
	if !errors.Is(err, ErrJournalRejected) || !strings.Contains(err.Error(), "No. Bukti sudah dipakai") {
		t.Errorf("got error %v, want vmedis' message", err)
	}
}

func TestCreateJournalEntry_RejectsNonPositiveAmount(t *testing.T) {
	c, gw := newJournalTestClient()

	if _, err := c.CreateJournalEntry(context.Background(), NewJournalEntry{Amount: 0}); err == nil {
		t.Error("got no error for a zero amount")
	}
	if len(gw.requests) != 0 {
		t.Errorf("got %d calls, want none", len(gw.requests))
	}
}

func TestFindJournalEntries(t *testing.T) {
	f := loadJournalFixture(t, "journal_list.json")
	c, gw := newJournalTestClient(f)

	day := time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local)
	got, err := c.FindJournalEntries(context.Background(), JournalEntryFilter{From: day, Until: day})
	if err != nil {
		t.Fatalf("FindJournalEntries: %v", err)
	}

	want := []JournalEntry{
		{ID: 1000001, VoucherNo: "JRLM2610310001", Time: time.Date(2026, 10, 31, 10, 0, 0, 0, time.Local), Description: "Gaji Karyawan Contoh Oktober 2026 (Snapshot #42)", Amount: 2500000, Username: "admin"},
		{ID: 1000002, VoucherNo: "JRLM2610310002", Time: time.Date(2026, 10, 31, 10, 0, 5, 0, time.Local), Description: "Pelunasan Utang Kasbon - Karyawan Contoh Oktober 2026 (Snapshot #42)", Amount: 150000, Username: "admin"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%+v\nwant\n%+v", got, want)
	}

	// The fixture is a page of 10 from the vmedis page; we page by 100. The
	// page adds app_jenis and uid only on some calls, we always send them.
	assertJournalRequest(t, gw, f, "limit", "app_jenis", "uid")
}

func TestDeleteJournalEntry(t *testing.T) {
	f := loadJournalFixture(t, "journal_delete.json")
	c, gw := newJournalTestClient(f)

	if err := c.DeleteJournalEntry(context.Background(), 1000001); err != nil {
		t.Fatalf("DeleteJournalEntry: %v", err)
	}

	assertJournalRequest(t, gw, f, "app_jenis", "uid")
}

func TestJournalCalls_NeedUserID(t *testing.T) {
	c := &Client{gateway: &routingGateway{}, appID: "v0000", reg: "ab12"}

	if _, err := c.ListJournalAccounts(context.Background()); err == nil {
		t.Error("got no error without a user id")
	}
}
