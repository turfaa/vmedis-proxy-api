package journal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"
	"golang.org/x/time/rate"
	"gorm.io/datatypes"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

const (
	MaxIdempotencyKeyLen = 100
	MaxDescriptionLen    = 500

	// staleAfter is how long a request can stay pending before it is taken
	// for one whose process died while sending it. It is longer than a send
	// can take: the lock wait plus two gateway calls.
	staleAfter = 5 * time.Minute

	lockWaitTimeout = 90 * time.Second
	sendTimeout     = 3 * time.Minute
)

var (
	ErrNotFound = errors.New("tidak ditemukan")
	// ErrInProgress is returned for a request whose key is being sent by
	// another request right now.
	ErrInProgress = errors.New("jurnal dengan idempotency key ini sedang diproses, coba lagi sebentar lagi")
	// ErrKeyReused is returned for a request whose key was used before for a
	// different entry.
	ErrKeyReused = errors.New("idempotency key sudah dipakai untuk jurnal lain")
	// ErrVmedis is returned when vmedis fails or can't be reached.
	ErrVmedis = errors.New("gagal menghubungi vmedis")
	// ErrOutcomeUnknown is returned with ErrVmedis when vmedis may have saved
	// the entry anyway. A retry with the same key finds it if it did.
	ErrOutcomeUnknown = errors.New("jurnal mungkin sudah tersimpan di vmedis")
)

// ValidationError is an invalid request.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type vmedisClient interface {
	ListJournalAccounts(ctx context.Context) ([]vmedisv2.JournalAccount, error)
	ReserveJournalVoucherNo(ctx context.Context, date time.Time) (string, error)
	CreateJournalEntry(ctx context.Context, entry vmedisv2.NewJournalEntry) (vmedisv2.CreatedJournalEntry, error)
	FindJournalEntries(ctx context.Context, filter vmedisv2.JournalEntryFilter) ([]vmedisv2.JournalEntry, error)
}

type requestStore interface {
	GetByKey(ctx context.Context, key string) (models.JournalEntryRequest, error)
	Insert(ctx context.Context, req *models.JournalEntryRequest) (bool, error)
	Claim(ctx context.Context, req *models.JournalEntryRequest, now time.Time) (bool, error)
	Save(ctx context.Context, req *models.JournalEntryRequest) error
}

type redisStore interface {
	AcquireCreateLock(ctx context.Context) (func(), error)
	GetAccounts(ctx context.Context) ([]vmedisv2.JournalAccount, error)
	SetAccounts(ctx context.Context, accounts []vmedisv2.JournalAccount) error
}

// Service creates journal entries in vmedis, each at most once per
// idempotency key.
type Service struct {
	db      requestStore
	redis   redisStore
	vmedis  vmedisClient
	limiter *rate.Limiter
	now     func() time.Time
}

// NewService creates a Service. Every vmedis call waits for limiter, which
// may be nil.
func NewService(db requestStore, redisDB redisStore, vmedis vmedisClient, limiter *rate.Limiter) *Service {
	return &Service{
		db:      db,
		redis:   redisDB,
		vmedis:  vmedis,
		limiter: limiter,
		now:     time.Now,
	}
}

// Create creates the journal entry, once per idempotency key: a request with
// a key that was used before returns the entry created then, with Replayed
// set. A request whose earlier attempt may have been saved by vmedis without
// an answer is first looked up in vmedis, by date, description and amount.
func (s *Service) Create(ctx context.Context, key string, req CreateRequest, email string) (CreateResponse, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return CreateResponse{}, &ValidationError{Message: "header Idempotency-Key wajib diisi"}
	}
	if utf8.RuneCountInString(key) > MaxIdempotencyKeyLen {
		return CreateResponse{}, &ValidationError{Message: fmt.Sprintf("Idempotency-Key maksimal %d karakter", MaxIdempotencyKeyLen)}
	}

	payload, err := parseCreateRequest(req)
	if err != nil {
		return CreateResponse{}, err
	}

	now := s.now()
	row := models.JournalEntryRequest{
		CreatedAt:      now,
		UpdatedAt:      now,
		IdempotencyKey: key,
		Status:         models.JournalEntryRequestPending,
		Attempts:       1,
		Request:        datatypes.NewJSONType(payload),
		CreatedByEmail: email,
	}

	inserted, err := s.db.Insert(ctx, &row)
	if err != nil {
		return CreateResponse{}, err
	}
	if inserted {
		return s.send(ctx, &row, payload)
	}

	row, err = s.db.GetByKey(ctx, key)
	if err != nil {
		return CreateResponse{}, err
	}

	if row.Request.Data() != payload {
		return CreateResponse{}, ErrKeyReused
	}

	// Whether the earlier attempt may have been saved by vmedis.
	mayExist := false
	switch row.Status {
	case models.JournalEntryRequestSucceeded:
		return responseOf(row, true), nil

	case models.JournalEntryRequestPending:
		if now.Sub(row.UpdatedAt) < staleAfter {
			return CreateResponse{}, ErrInProgress
		}
		// Its process died while sending it.
		mayExist = true

	case models.JournalEntryRequestUnknown:
		mayExist = true
	}

	claimed, err := s.db.Claim(ctx, &row, now)
	if err != nil {
		return CreateResponse{}, err
	}
	if !claimed {
		return CreateResponse{}, ErrInProgress
	}

	if mayExist {
		existing, found, err := s.findExisting(ctx, payload)
		if err != nil {
			s.fail(ctx, &row, models.JournalEntryRequestUnknown, err)
			return CreateResponse{}, fmt.Errorf("%w: %w: gagal mengecek: %s", ErrVmedis, ErrOutcomeUnknown, err)
		}

		if found {
			log.Printf("Journal entry request %q was saved in vmedis as %s (%d) without an answer", key, existing.VoucherNo, existing.ID)
			s.succeed(ctx, &row, existing.ID, existing.VoucherNo)
			return responseOf(row, true), nil
		}
	}

	return s.send(ctx, &row, payload)
}

// send creates the entry of row, which this process claimed, in vmedis.
func (s *Service) send(ctx context.Context, row *models.JournalEntryRequest, payload models.JournalEntryRequestPayload) (CreateResponse, error) {
	// Once started, a send runs to completion even when the client goes
	// away, so its outcome is always recorded.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), sendTimeout)
	defer cancel()

	lockCtx, cancelLock := context.WithTimeout(ctx, lockWaitTimeout)
	release, err := s.redis.AcquireCreateLock(lockCtx)
	cancelLock()
	if err != nil {
		s.fail(ctx, row, models.JournalEntryRequestFailed, err)
		return CreateResponse{}, err
	}
	defer release()

	debitID, creditID, err := s.resolveAccounts(ctx, payload.DebitAccountCode, payload.CreditAccountCode)
	if err != nil {
		s.fail(ctx, row, models.JournalEntryRequestFailed, err)
		return CreateResponse{}, err
	}

	date, _ := time.ParseInLocation(time.DateOnly, payload.Date, time.Local)

	if err := s.wait(ctx); err != nil {
		s.fail(ctx, row, models.JournalEntryRequestFailed, err)
		return CreateResponse{}, err
	}
	voucherNo, err := s.vmedis.ReserveJournalVoucherNo(ctx, date)
	if err != nil {
		s.fail(ctx, row, models.JournalEntryRequestFailed, err)
		return CreateResponse{}, fmt.Errorf("%w: %s", ErrVmedis, err)
	}

	entry := vmedisv2.NewJournalEntry{
		Date:            date,
		VoucherNo:       voucherNo,
		Description:     payload.Description,
		DebitAccountID:  debitID,
		CreditAccountID: creditID,
		Amount:          payload.Amount,
	}

	// Recorded before sending, so the audit trail has it whatever happens.
	row.VmedisPayload = datatypes.NewJSONType(&models.JournalEntryVmedisPayload{
		Date:            payload.Date,
		VoucherNo:       voucherNo,
		Description:     entry.Description,
		DebitAccountID:  debitID,
		CreditAccountID: creditID,
		Amount:          entry.Amount,
	})
	row.UpdatedAt = s.now()
	if err := s.db.Save(ctx, row); err != nil {
		// Nothing was sent. The row stays pending until it is stale, and is
		// then looked up before being sent again.
		return CreateResponse{}, err
	}

	if err := s.wait(ctx); err != nil {
		s.fail(ctx, row, models.JournalEntryRequestFailed, err)
		return CreateResponse{}, err
	}
	created, err := s.vmedis.CreateJournalEntry(ctx, entry)
	if err != nil {
		if errors.Is(err, vmedisv2.ErrJournalRejected) {
			s.fail(ctx, row, models.JournalEntryRequestFailed, err)
			return CreateResponse{}, fmt.Errorf("%w: %s", ErrVmedis, err)
		}

		s.fail(ctx, row, models.JournalEntryRequestUnknown, err)
		return CreateResponse{}, fmt.Errorf("%w: %w: %s", ErrVmedis, ErrOutcomeUnknown, err)
	}

	s.succeed(ctx, row, created.ID, created.VoucherNo)
	return responseOf(*row, false), nil
}

// Find lists the active journal entries dated date whose description contains
// query, ignoring case. An empty query lists every entry of the date.
func (s *Service) Find(ctx context.Context, date string, query string) (FindResponse, error) {
	day, err := time.ParseInLocation(time.DateOnly, strings.TrimSpace(date), time.Local)
	if err != nil {
		return FindResponse{}, &ValidationError{Message: fmt.Sprintf("tanggal %q tidak valid, harus YYYY-MM-DD", date)}
	}

	entries, err := s.findOn(ctx, day)
	if err != nil {
		return FindResponse{}, fmt.Errorf("%w: %s", ErrVmedis, err)
	}

	query = strings.ToLower(strings.TrimSpace(query))
	res := FindResponse{Entries: []Entry{}}
	for _, e := range entries {
		if query != "" && !strings.Contains(strings.ToLower(e.Description), query) {
			continue
		}

		res.Entries = append(res.Entries, Entry{
			VmedisID:    e.ID,
			VoucherNo:   e.VoucherNo,
			Time:        e.Time,
			Description: e.Description,
			Amount:      strconv.FormatFloat(e.Amount, 'f', -1, 64),
			Username:    e.Username,
		})
	}

	return res, nil
}

// findExisting looks for an active entry in vmedis with the payload's date,
// description and amount. (The listing has no accounts.)
func (s *Service) findExisting(ctx context.Context, payload models.JournalEntryRequestPayload) (vmedisv2.JournalEntry, bool, error) {
	day, _ := time.ParseInLocation(time.DateOnly, payload.Date, time.Local)

	entries, err := s.findOn(ctx, day)
	if err != nil {
		return vmedisv2.JournalEntry{}, false, err
	}

	for _, e := range entries {
		if strings.TrimSpace(e.Description) == payload.Description && e.Amount == float64(payload.Amount) {
			return e, true, nil
		}
	}

	return vmedisv2.JournalEntry{}, false, nil
}

// findOn lists the entries dated day. The keterangan filter of vmedis is not
// used: whether and how it matches is unverified, and a day has few entries,
// so they are matched here instead.
func (s *Service) findOn(ctx context.Context, day time.Time) ([]vmedisv2.JournalEntry, error) {
	if err := s.wait(ctx); err != nil {
		return nil, err
	}

	return s.vmedis.FindJournalEntries(ctx, vmedisv2.JournalEntryFilter{From: day, Until: day})
}

// resolveAccounts returns the vmedis IDs of the account codes, refreshing the
// cached account list when a code isn't in it.
func (s *Service) resolveAccounts(ctx context.Context, codes ...string) (debitID, creditID int, err error) {
	accounts, err := s.redis.GetAccounts(ctx)
	if err != nil {
		log.Printf("Getting cached journal accounts: %s", err)
	}

	ids, missing := accountIDs(accounts, codes)
	if missing != "" {
		if err := s.wait(ctx); err != nil {
			return 0, 0, err
		}
		accounts, err = s.vmedis.ListJournalAccounts(ctx)
		if err != nil {
			return 0, 0, fmt.Errorf("%w: memuat daftar akun: %s", ErrVmedis, err)
		}

		if err := s.redis.SetAccounts(ctx, accounts); err != nil {
			log.Printf("Caching journal accounts: %s", err)
		}

		ids, missing = accountIDs(accounts, codes)
		if missing != "" {
			return 0, 0, &ValidationError{Message: fmt.Sprintf("akun %s tidak ada di vmedis", missing)}
		}
	}

	return ids[0], ids[1], nil
}

// accountIDs returns the IDs of the codes, or the first code without one.
func accountIDs(accounts []vmedisv2.JournalAccount, codes []string) ([]int, string) {
	byCode := make(map[string]int, len(accounts))
	for _, a := range accounts {
		byCode[a.Code] = a.ID
	}

	ids := make([]int, len(codes))
	for i, code := range codes {
		id, ok := byCode[code]
		if !ok {
			return nil, code
		}
		ids[i] = id
	}

	return ids, ""
}

func (s *Service) wait(ctx context.Context) error {
	if s.limiter == nil {
		return nil
	}

	return s.limiter.Wait(ctx)
}

// fail records that the attempt of row ended with err, in status.
func (s *Service) fail(ctx context.Context, row *models.JournalEntryRequest, status string, err error) {
	message := err.Error()
	row.Status = status
	row.Error = &message
	row.UpdatedAt = s.now()

	if saveErr := s.db.Save(context.WithoutCancel(ctx), row); saveErr != nil {
		log.Printf("Recording journal entry request %q as %s (%s): %s", row.IdempotencyKey, status, message, saveErr)
	}
}

// succeed records that the entry of row was saved in vmedis.
func (s *Service) succeed(ctx context.Context, row *models.JournalEntryRequest, vmedisID int, voucherNo string) {
	row.Status = models.JournalEntryRequestSucceeded
	row.VmedisID = &vmedisID
	row.VoucherNo = &voucherNo
	row.Error = nil
	row.UpdatedAt = s.now()

	if err := s.db.Save(context.WithoutCancel(ctx), row); err != nil {
		// The entry exists in vmedis. The row stays pending until it is stale,
		// and is then found in vmedis instead of being sent again.
		log.Printf("Recording journal entry request %q as saved in vmedis as %s (%d): %s", row.IdempotencyKey, voucherNo, vmedisID, err)
	}
}

func responseOf(row models.JournalEntryRequest, replayed bool) CreateResponse {
	res := CreateResponse{Replayed: replayed}
	if row.VmedisID != nil {
		res.VmedisID = *row.VmedisID
	}
	if row.VoucherNo != nil {
		res.VoucherNo = *row.VoucherNo
	}
	return res
}

func parseCreateRequest(req CreateRequest) (models.JournalEntryRequestPayload, error) {
	invalid := func(format string, args ...any) (models.JournalEntryRequestPayload, error) {
		return models.JournalEntryRequestPayload{}, &ValidationError{Message: fmt.Sprintf(format, args...)}
	}

	date := strings.TrimSpace(req.Date)
	if _, err := time.ParseInLocation(time.DateOnly, date, time.Local); err != nil {
		return invalid("tanggal %q tidak valid, harus YYYY-MM-DD", req.Date)
	}

	description := strings.TrimSpace(req.Description)
	if description == "" {
		return invalid("keterangan wajib diisi")
	}
	if utf8.RuneCountInString(description) > MaxDescriptionLen {
		return invalid("keterangan maksimal %d karakter", MaxDescriptionLen)
	}

	credit, debit := strings.TrimSpace(req.CreditAccountCode), strings.TrimSpace(req.DebitAccountCode)
	if credit == "" || debit == "" {
		return invalid("akun kredit dan akun debit wajib diisi")
	}
	if credit == debit {
		return invalid("akun kredit dan akun debit tidak boleh sama")
	}

	amount, err := decimal.NewFromString(strings.TrimSpace(req.Amount))
	if err != nil {
		return invalid("nominal %q tidak valid", req.Amount)
	}
	if !amount.IsPositive() {
		return invalid("nominal harus lebih dari 0")
	}
	if !amount.IsInteger() {
		return invalid("nominal harus bilangan bulat Rupiah, bukan %s", amount)
	}
	if amount.GreaterThan(decimal.NewFromInt(1e15)) {
		return invalid("nominal terlalu besar")
	}

	return models.JournalEntryRequestPayload{
		Date:              date,
		Description:       description,
		CreditAccountCode: credit,
		DebitAccountCode:  debit,
		Amount:            amount.IntPart(),
	}, nil
}
