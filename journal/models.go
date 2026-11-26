// Package journal creates and looks up journal entries (jurnal keuangan) in
// vmedis, e.g. the salary entries HRIS posts from salary snapshots.
package journal

import "time"

// CreateRequest is a journal entry to create: Amount moves from the credit
// account to the debit account.
type CreateRequest struct {
	// Date is YYYY-MM-DD.
	Date        string `json:"date"`
	Description string `json:"description"`
	// CreditAccountCode and DebitAccountCode are account codes (akkode),
	// e.g. 111004.
	CreditAccountCode string `json:"creditAccountCode"`
	DebitAccountCode  string `json:"debitAccountCode"`
	// Amount is a decimal string in whole Rupiah, e.g. "2500000" or
	// "2500000.00".
	Amount string `json:"amount"`
}

// CreateResponse identifies the created journal entry.
type CreateResponse struct {
	// VoucherNo is the vmedis No. Bukti, e.g. JRLM2610310001.
	VoucherNo string `json:"voucherNo"`
	VmedisID  int    `json:"vmedisId"`
	// Replayed is true when the idempotency key was used before, and the
	// entry is the one created then.
	Replayed bool `json:"replayed"`
}

// Entry is a journal entry in vmedis.
type Entry struct {
	VmedisID    int       `json:"vmedisId"`
	VoucherNo   string    `json:"voucherNo"`
	Time        time.Time `json:"time"`
	Description string    `json:"description"`
	// Amount is a decimal string.
	Amount   string `json:"amount"`
	Username string `json:"username"`
}

// FindResponse is the journal entries found.
type FindResponse struct {
	Entries []Entry `json:"entries"`
}
