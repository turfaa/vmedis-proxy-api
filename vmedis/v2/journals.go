package vmedisv2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// The calls below are the ones the vmedis "Data Jurnal Keuangan" page
// (/jurnal-v2) makes. They need the user ID (see WithUserID), since vmedis
// records the entries under that user.

// ErrJournalRejected wraps the errors of jurnal/* calls that vmedis answered
// with success: false. Such a call wrote nothing, unlike one that failed in
// transit, whose outcome is unknown.
var ErrJournalRejected = errors.New("vmedis rejected the request")

// JournalAccount is an account (akun) journal entries are debited or credited
// to, e.g. 111004 / Bank OCBC.
type JournalAccount struct {
	// ID is the internal vmedis ID that journal entries refer to the account by.
	ID   int
	Code string
	Name string
}

// ListJournalAccounts gets every account journal entries can use.
func (c *Client) ListJournalAccounts(ctx context.Context) ([]JournalAccount, error) {
	var res struct {
		journalResponse
		Data []struct {
			ID   number `json:"id"`
			Code string `json:"akkode"`
			Name string `json:"aknama"`
		} `json:"data"`
	}
	if err := c.callJournal(ctx, "jurnal/akun-options", nil, &res, false); err != nil {
		return nil, err
	}
	if err := res.err(); err != nil {
		return nil, fmt.Errorf("list journal accounts: %w", err)
	}

	accounts := make([]JournalAccount, len(res.Data))
	for i, row := range res.Data {
		accounts[i] = JournalAccount{
			ID:   int(row.ID),
			Code: strings.TrimSpace(row.Code),
			Name: strings.TrimSpace(row.Name),
		}
	}

	return accounts, nil
}

// ReserveJournalVoucherNo gets the next No. Bukti (e.g. JRLM2610310001) for a
// journal entry dated date. It does not save anything: the number is used by
// the next CreateJournalEntry for that date, so callers must not let two
// reserve/create pairs interleave.
func (c *Client) ReserveJournalVoucherNo(ctx context.Context, date time.Time) (string, error) {
	isoDate := date.Format(time.DateOnly)

	var res struct {
		journalResponse
		VoucherNo string `json:"jrlno"`
	}
	if err := c.callJournal(ctx, "jurnal/get-kode", map[string]any{
		"tanggal":    isoDate,
		"jrltanggal": isoDate,
	}, &res, false); err != nil {
		return "", err
	}
	if err := res.err(); err != nil {
		return "", fmt.Errorf("reserve journal voucher number: %w", err)
	}

	voucherNo := strings.TrimSpace(res.VoucherNo)
	if voucherNo == "" {
		return "", fmt.Errorf("reserve journal voucher number: vmedis returned no number")
	}

	return voucherNo, nil
}

// NewJournalEntry is a journal entry to create: Amount moves from the credit
// account to the debit account.
type NewJournalEntry struct {
	Date time.Time
	// VoucherNo is the No. Bukti from ReserveJournalVoucherNo.
	VoucherNo       string
	Description     string
	DebitAccountID  int
	CreditAccountID int
	// Amount is in whole Rupiah.
	Amount int64
}

// CreatedJournalEntry identifies a journal entry created in vmedis.
type CreatedJournalEntry struct {
	// ID is vmedis' internal ID (jrlid).
	ID int
	// VoucherNo is the No. Bukti (jrlno).
	VoucherNo string
}

// CreateJournalEntry saves a journal entry. It is sent once, without retries:
// when it fails, the entry may or may not have been saved, which callers can
// check with FindJournalEntries.
func (c *Client) CreateJournalEntry(ctx context.Context, entry NewJournalEntry) (CreatedJournalEntry, error) {
	if entry.Amount <= 0 {
		return CreatedJournalEntry{}, fmt.Errorf("create journal entry: amount must be positive, got %d", entry.Amount)
	}

	var res struct {
		journalResponse
		ID        number `json:"jrlid"`
		VoucherNo string `json:"jrlno"`
	}
	if err := c.callJournal(ctx, "jurnal/create", map[string]any{
		"jrltanggal":    entry.Date.Format(time.DateOnly),
		"jrlno":         entry.VoucherNo,
		"jrlketerangan": entry.Description,
		// No Jurnal Singkat preset: the accounts are set directly.
		"jsid":          0,
		"jrlkodedebit":  entry.DebitAccountID,
		"jrlkodekredit": entry.CreditAccountID,
		"djdebit":       entry.DebitAccountID,
		"djkredit":      entry.CreditAccountID,
		"jrlnominal":    entry.Amount,
		// The page sends the current time; the date above is what counts.
		"time": time.Now().Format(time.TimeOnly),
	}, &res, true); err != nil {
		return CreatedJournalEntry{}, err
	}
	if err := res.err(); err != nil {
		return CreatedJournalEntry{}, fmt.Errorf("create journal entry: %w", err)
	}

	created := CreatedJournalEntry{ID: int(res.ID), VoucherNo: strings.TrimSpace(res.VoucherNo)}
	if created.ID == 0 {
		return CreatedJournalEntry{}, fmt.Errorf("create journal entry: vmedis returned no jrlid")
	}
	if created.VoucherNo == "" {
		created.VoucherNo = entry.VoucherNo
	}

	return created, nil
}

// JournalEntry is an active journal entry as listed by vmedis. The listing
// has no account fields.
type JournalEntry struct {
	ID          int
	VoucherNo   string
	Time        time.Time
	Description string
	Amount      float64
	Username    string
}

// JournalEntryFilter filters FindJournalEntries.
type JournalEntryFilter struct {
	// From and Until are the first and last dates listed, inclusive.
	From, Until time.Time
	// Description, when set, is passed to vmedis' keterangan filter.
	Description string
}

// journalEntriesPageSize is the number of journal entries fetched per page.
const journalEntriesPageSize = 100

// FindJournalEntries lists the active journal entries in the filter's date
// range.
func (c *Client) FindJournalEntries(ctx context.Context, filter JournalEntryFilter) ([]JournalEntry, error) {
	var entries []JournalEntry
	for offset := 0; ; offset += journalEntriesPageSize {
		var res struct {
			journalResponse
			Data []struct {
				ID          number `json:"jrlid"`
				Date        string `json:"jrltanggal"`
				VoucherNo   string `json:"jrlno"`
				Description string `json:"jrlketerangan"`
				Username    string `json:"username"`
				Amount      number `json:"nominal"`
			} `json:"data"`
		}
		if err := c.callJournal(ctx, "jurnal/list", map[string]any{
			"cari":          4,
			"tanggalawal":   formatJournalListDate(filter.From),
			"tanggalakhir":  formatJournalListDate(filter.Until),
			"bulan":         filter.From.Format("Jan 2006"),
			"tahun":         filter.From.Format("2006"),
			"jrlno":         "",
			"jrlketerangan": filter.Description,
			"username":      "",
			"jrlaktif":      "1",
			"limit":         journalEntriesPageSize,
			"offset":        offset,
		}, &res, false); err != nil {
			return nil, err
		}
		if err := res.err(); err != nil {
			return nil, fmt.Errorf("list journal entries: %w", err)
		}

		for _, row := range res.Data {
			t, err := time.ParseInLocation(time.DateTime, strings.TrimSpace(row.Date), time.Local)
			if err != nil {
				log.Printf("Journal entry %d has an unparseable date %q: %s", int(row.ID), row.Date, err)
			}

			entries = append(entries, JournalEntry{
				ID:          int(row.ID),
				VoucherNo:   strings.TrimSpace(row.VoucherNo),
				Time:        t,
				Description: row.Description,
				Amount:      float64(row.Amount),
				Username:    strings.TrimSpace(row.Username),
			})
		}

		if len(res.Data) < journalEntriesPageSize {
			break
		}
	}

	return entries, nil
}

// DeleteJournalEntry deletes the journal entry with vmedis ID id.
func (c *Client) DeleteJournalEntry(ctx context.Context, id int) error {
	var res journalResponse
	if err := c.callJournal(ctx, "jurnal/delete", map[string]any{"jrlid": id}, &res, true); err != nil {
		return err
	}
	if err := res.err(); err != nil {
		return fmt.Errorf("delete journal entry %d: %w", id, err)
	}

	return nil
}

// formatJournalListDate formats a date the way jurnal/list takes it, e.g.
// "31 Oct 2026". The month is always the English abbreviation.
func formatJournalListDate(t time.Time) string {
	return t.Format("2 Jan 2006")
}

// callJournal calls a jurnal/* endpoint, adding the parameters every one of
// them takes.
func (c *Client) callJournal(ctx context.Context, targetUrl string, payload map[string]any, responseTarget any, once bool) error {
	if c.userID == 0 {
		return fmt.Errorf("vmedis v2 user id must be configured to call %s", targetUrl)
	}

	params := make(map[string]any, len(payload)+3)
	for k, v := range payload {
		params[k] = v
	}
	params["mn_jenis"] = 1
	params["app_jenis"] = 1
	params["uid"] = c.userID

	if once {
		return c.callReportOnce(ctx, targetUrl, params, responseTarget)
	}
	return c.callReport(ctx, targetUrl, params, responseTarget)
}

// journalResponse is the envelope of the jurnal/* responses.
type journalResponse struct {
	Success bool `json:"success"`
	// Message is set on writes, e.g. "Data tersimpan", and on failures.
	Message json.RawMessage `json:"message"`
}

func (r journalResponse) err() error {
	if r.Success {
		return nil
	}

	var message string
	if err := json.Unmarshal(r.Message, &message); err != nil {
		message = string(r.Message)
	}
	if message == "" {
		message = "no message"
	}

	return fmt.Errorf("%w: %s", ErrJournalRejected, message)
}
