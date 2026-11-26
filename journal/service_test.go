package journal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gorm.io/datatypes"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

type fakeVmedis struct {
	mu sync.Mutex

	accounts  []vmedisv2.JournalAccount
	entries   []vmedisv2.JournalEntry
	createErr error
	// saveDespiteErr saves the entry even when createErr is returned, like a
	// timeout after vmedis saved it.
	saveDespiteErr bool

	accountCalls, reserveCalls, createCalls, findCalls int
	created                                            []vmedisv2.NewJournalEntry
}

func (f *fakeVmedis) ListJournalAccounts(context.Context) ([]vmedisv2.JournalAccount, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.accountCalls++
	return f.accounts, nil
}

func (f *fakeVmedis) ReserveJournalVoucherNo(_ context.Context, date time.Time) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reserveCalls++
	return fmt.Sprintf("JRLM%s%04d", date.Format("060102"), len(f.entries)+1), nil
}

func (f *fakeVmedis) CreateJournalEntry(_ context.Context, e vmedisv2.NewJournalEntry) (vmedisv2.CreatedJournalEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createCalls++

	if f.createErr != nil && !f.saveDespiteErr {
		return vmedisv2.CreatedJournalEntry{}, f.createErr
	}

	f.created = append(f.created, e)
	id := 1000 + len(f.entries)
	f.entries = append(f.entries, vmedisv2.JournalEntry{ID: id, VoucherNo: e.VoucherNo, Time: e.Date, Description: e.Description, Amount: float64(e.Amount)})

	if f.createErr != nil {
		return vmedisv2.CreatedJournalEntry{}, f.createErr
	}
	return vmedisv2.CreatedJournalEntry{ID: id, VoucherNo: e.VoucherNo}, nil
}

func (f *fakeVmedis) FindJournalEntries(_ context.Context, filter vmedisv2.JournalEntryFilter) ([]vmedisv2.JournalEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.findCalls++

	var res []vmedisv2.JournalEntry
	for _, e := range f.entries {
		if !e.Time.Before(filter.From) && e.Time.Before(filter.Until.AddDate(0, 0, 1)) {
			res = append(res, e)
		}
	}
	return res, nil
}

type fakeRedis struct {
	mu       sync.Mutex
	accounts []vmedisv2.JournalAccount
}

func (f *fakeRedis) AcquireCreateLock(context.Context) (func(), error) {
	f.mu.Lock()
	return f.mu.Unlock, nil
}

func (f *fakeRedis) GetAccounts(context.Context) ([]vmedisv2.JournalAccount, error) {
	return f.accounts, nil
}

func (f *fakeRedis) SetAccounts(_ context.Context, accounts []vmedisv2.JournalAccount) error {
	f.accounts = accounts
	return nil
}

type testEnv struct {
	service *Service
	db      *Database
	vmedis  *fakeVmedis
	redis   *fakeRedis
	now     time.Time
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	gormDB, err := database.SqliteDB(t.TempDir() + "/test.db?_pragma=busy_timeout(10000)")
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	env := &testEnv{
		db: NewDatabase(gormDB),
		vmedis: &fakeVmedis{accounts: []vmedisv2.JournalAccount{
			{ID: 618, Code: "111004", Name: "Bank OCBC"},
			{ID: 592, Code: "171001", Name: "Piutang Pegawai"},
			{ID: 377, Code: "610002", Name: "Gaji Karyawan"},
		}},
		redis: &fakeRedis{},
		now:   time.Date(2026, 11, 26, 10, 0, 0, 0, time.Local),
	}
	env.service = NewService(env.db, env.redis, env.vmedis, nil)
	env.service.now = func() time.Time { return env.now }

	return env
}

var salaryRequest = CreateRequest{
	Date:              "2026-10-31",
	Description:       "Gaji Karyawan Contoh Oktober 2026 (Snapshot #42, diposting oleh Admin)",
	CreditAccountCode: "111004",
	DebitAccountCode:  "610002",
	Amount:            "2500000.00",
}

const salaryKey = "hris-snapshot-42-salary-0"

func (env *testEnv) row(t *testing.T, key string) models.JournalEntryRequest {
	t.Helper()

	row, err := env.db.GetByKey(context.Background(), key)
	if err != nil {
		t.Fatalf("GetByKey: %v", err)
	}
	return row
}

func TestCreate(t *testing.T) {
	env := newTestEnv(t)

	res, err := env.service.Create(context.Background(), salaryKey, salaryRequest, "admin@example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if res != (CreateResponse{VoucherNo: "JRLM2610310001", VmedisID: 1000}) {
		t.Errorf("got %+v", res)
	}

	want := vmedisv2.NewJournalEntry{
		Date:            time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local),
		VoucherNo:       "JRLM2610310001",
		Description:     salaryRequest.Description,
		DebitAccountID:  377,
		CreditAccountID: 618,
		Amount:          2500000,
	}
	if len(env.vmedis.created) != 1 || env.vmedis.created[0] != want {
		t.Errorf("got created %+v, want %+v", env.vmedis.created, want)
	}

	row := env.row(t, salaryKey)
	if row.Status != models.JournalEntryRequestSucceeded || *row.VoucherNo != "JRLM2610310001" || *row.VmedisID != 1000 || row.CreatedByEmail != "admin@example.com" {
		t.Errorf("got row %+v", row)
	}
	if p := row.VmedisPayload.Data(); p == nil || p.DebitAccountID != 377 || p.VoucherNo != "JRLM2610310001" {
		t.Errorf("got vmedis payload %+v", p)
	}

	// The account list is cached.
	if _, err := env.service.Create(context.Background(), "other", CreateRequest{
		Date: "2026-10-31", Description: "Pelunasan", CreditAccountCode: "171001", DebitAccountCode: "111004", Amount: "1",
	}, ""); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if env.vmedis.accountCalls != 1 {
		t.Errorf("got %d account list calls, want 1", env.vmedis.accountCalls)
	}
}

func TestCreate_ReplaysKey(t *testing.T) {
	env := newTestEnv(t)

	first, err := env.service.Create(context.Background(), salaryKey, salaryRequest, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	second, err := env.service.Create(context.Background(), salaryKey, salaryRequest, "")
	if err != nil {
		t.Fatalf("Create again: %v", err)
	}

	if !second.Replayed || second.VoucherNo != first.VoucherNo || second.VmedisID != first.VmedisID {
		t.Errorf("got %+v, want a replay of %+v", second, first)
	}
	if env.vmedis.createCalls != 1 {
		t.Errorf("got %d creates, want 1", env.vmedis.createCalls)
	}
}

func TestCreate_RejectsReusedKey(t *testing.T) {
	env := newTestEnv(t)

	if _, err := env.service.Create(context.Background(), salaryKey, salaryRequest, ""); err != nil {
		t.Fatalf("Create: %v", err)
	}

	other := salaryRequest
	other.Amount = "1"
	if _, err := env.service.Create(context.Background(), salaryKey, other, ""); !errors.Is(err, ErrKeyReused) {
		t.Errorf("got %v, want ErrKeyReused", err)
	}
}

func TestCreate_UnknownOutcomeIsLookedUpBeforeResending(t *testing.T) {
	env := newTestEnv(t)
	env.vmedis.createErr = errors.New("timeout")
	env.vmedis.saveDespiteErr = true

	if _, err := env.service.Create(context.Background(), salaryKey, salaryRequest, ""); !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("got %v, want ErrOutcomeUnknown", err)
	}
	if row := env.row(t, salaryKey); row.Status != models.JournalEntryRequestUnknown {
		t.Fatalf("got status %s, want unknown", row.Status)
	}

	env.vmedis.createErr = nil
	res, err := env.service.Create(context.Background(), salaryKey, salaryRequest, "")
	if err != nil {
		t.Fatalf("retry: %v", err)
	}

	if env.vmedis.createCalls != 1 {
		t.Errorf("got %d creates, want 1: the retry must find the saved entry", env.vmedis.createCalls)
	}
	if !res.Replayed || res.VoucherNo != "JRLM2610310001" || res.VmedisID != 1000 {
		t.Errorf("got %+v", res)
	}
	if row := env.row(t, salaryKey); row.Status != models.JournalEntryRequestSucceeded {
		t.Errorf("got status %s, want succeeded", row.Status)
	}
}

func TestCreate_UnknownOutcomeNotSavedIsResent(t *testing.T) {
	env := newTestEnv(t)
	env.vmedis.createErr = errors.New("connection reset")

	if _, err := env.service.Create(context.Background(), salaryKey, salaryRequest, ""); !errors.Is(err, ErrVmedis) {
		t.Fatalf("got %v, want ErrVmedis", err)
	}

	env.vmedis.createErr = nil
	if _, err := env.service.Create(context.Background(), salaryKey, salaryRequest, ""); err != nil {
		t.Fatalf("retry: %v", err)
	}

	if env.vmedis.findCalls != 1 || env.vmedis.createCalls != 2 || len(env.vmedis.entries) != 1 {
		t.Errorf("got %d finds, %d creates, %d entries; want 1, 2, 1", env.vmedis.findCalls, env.vmedis.createCalls, len(env.vmedis.entries))
	}
}

func TestCreate_RejectedIsResentWithoutLookup(t *testing.T) {
	env := newTestEnv(t)
	env.vmedis.createErr = fmt.Errorf("create journal entry: %w: gagal", vmedisv2.ErrJournalRejected)

	_, err := env.service.Create(context.Background(), salaryKey, salaryRequest, "")
	if !errors.Is(err, ErrVmedis) || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("got %v, want ErrVmedis with a known outcome", err)
	}
	if row := env.row(t, salaryKey); row.Status != models.JournalEntryRequestFailed || row.Error == nil {
		t.Fatalf("got row %+v, want failed with the error", row)
	}

	env.vmedis.createErr = nil
	if _, err := env.service.Create(context.Background(), salaryKey, salaryRequest, ""); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if env.vmedis.findCalls != 0 {
		t.Errorf("got %d finds, want none", env.vmedis.findCalls)
	}
}

func TestCreate_PendingIsInProgressUntilStale(t *testing.T) {
	env := newTestEnv(t)

	// A request whose process died right after it was claimed, having saved
	// the entry in vmedis.
	row := models.JournalEntryRequest{
		CreatedAt: env.now, UpdatedAt: env.now, IdempotencyKey: salaryKey,
		Status: models.JournalEntryRequestPending, Attempts: 1,
	}
	payload, _ := parseCreateRequest(salaryRequest)
	row.Request = datatypes.NewJSONType(payload)
	if _, err := env.db.Insert(context.Background(), &row); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	env.vmedis.entries = []vmedisv2.JournalEntry{{ID: 7, VoucherNo: "JRLM2610310007", Time: time.Date(2026, 10, 31, 10, 0, 0, 0, time.Local), Description: payload.Description, Amount: 2500000}}

	if _, err := env.service.Create(context.Background(), salaryKey, salaryRequest, ""); !errors.Is(err, ErrInProgress) {
		t.Fatalf("got %v, want ErrInProgress", err)
	}

	env.now = env.now.Add(staleAfter)
	res, err := env.service.Create(context.Background(), salaryKey, salaryRequest, "")
	if err != nil {
		t.Fatalf("Create when stale: %v", err)
	}
	if res.VmedisID != 7 || env.vmedis.createCalls != 0 {
		t.Errorf("got %+v after %d creates, want the existing entry", res, env.vmedis.createCalls)
	}
}

func TestCreate_Validation(t *testing.T) {
	env := newTestEnv(t)

	for name, tc := range map[string]struct {
		key string
		req CreateRequest
	}{
		"no key":          {"", salaryRequest},
		"fractional":      {"k", with(func(r *CreateRequest) { r.Amount = "10.5" })},
		"zero":            {"k", with(func(r *CreateRequest) { r.Amount = "0" })},
		"bad date":        {"k", with(func(r *CreateRequest) { r.Date = "31-10-2026" })},
		"no description":  {"k", with(func(r *CreateRequest) { r.Description = " " })},
		"same accounts":   {"k", with(func(r *CreateRequest) { r.DebitAccountCode = r.CreditAccountCode })},
		"missing account": {"k", with(func(r *CreateRequest) { r.DebitAccountCode = "999999" })},
	} {
		t.Run(name, func(t *testing.T) {
			var validationErr *ValidationError
			if _, err := env.service.Create(context.Background(), tc.key, tc.req, ""); !errors.As(err, &validationErr) {
				t.Errorf("got %v, want a ValidationError", err)
			}
		})
	}

	if env.vmedis.createCalls != 0 {
		t.Errorf("got %d creates, want none", env.vmedis.createCalls)
	}
}

func TestFind(t *testing.T) {
	env := newTestEnv(t)
	day := time.Date(2026, 10, 31, 10, 0, 0, 0, time.Local)
	env.vmedis.entries = []vmedisv2.JournalEntry{
		{ID: 1, VoucherNo: "A", Time: day, Description: "Gaji X (Snapshot #42, diposting oleh Y)", Amount: 10},
		{ID: 2, VoucherNo: "B", Time: day, Description: "Beli ATK", Amount: 2.5},
		{ID: 3, VoucherNo: "C", Time: day.AddDate(0, 0, 1), Description: "Gaji X (Snapshot #42)", Amount: 10},
	}

	res, err := env.service.Find(context.Background(), "2026-10-31", "snapshot #42")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(res.Entries) != 1 || res.Entries[0].VmedisID != 1 || res.Entries[0].Amount != "10" {
		t.Errorf("got %+v", res.Entries)
	}

	var validationErr *ValidationError
	if _, err := env.service.Find(context.Background(), "", ""); !errors.As(err, &validationErr) {
		t.Errorf("got %v, want a ValidationError", err)
	}
}

func with(edit func(*CreateRequest)) CreateRequest {
	r := salaryRequest
	edit(&r)
	return r
}
