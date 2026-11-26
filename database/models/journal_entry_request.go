package models

import (
	"time"

	"gorm.io/datatypes"
)

// Statuses of a JournalEntryRequest.
const (
	// JournalEntryRequestPending is a request being sent to vmedis, or one
	// whose process died while sending it.
	JournalEntryRequestPending = "pending"
	// JournalEntryRequestSucceeded is a request whose entry vmedis saved.
	JournalEntryRequestSucceeded = "succeeded"
	// JournalEntryRequestFailed is a request vmedis saved nothing for, so it
	// can be sent again.
	JournalEntryRequestFailed = "failed"
	// JournalEntryRequestUnknown is a request whose create call failed in
	// transit (e.g. timed out), so vmedis may or may not have saved it. It is
	// looked up in vmedis before being sent again.
	JournalEntryRequestUnknown = "unknown"
)

// JournalEntryRequest is a request to create a journal entry (jurnal
// keuangan) in vmedis, kept for idempotency and audit.
type JournalEntryRequest struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index;not null"`
	UpdatedAt time.Time `gorm:"not null"`

	// IdempotencyKey is chosen by the client, e.g. hris-snapshot-42-salary-0.
	// A request with a key that was used before returns that request's entry
	// instead of creating another one.
	IdempotencyKey string `gorm:"uniqueIndex;not null"`

	Status string `gorm:"index;not null"`
	// Attempts counts the times the request was claimed to be sent. A claim
	// only succeeds from the count it read, so two processes never send the
	// same request at once.
	Attempts int `gorm:"not null;default:0"`

	// Request is the entry as requested.
	Request datatypes.JSONType[JournalEntryRequestPayload] `gorm:"not null"`
	// VmedisPayload is what was last sent to jurnal/create, when it was sent.
	VmedisPayload datatypes.JSONType[*JournalEntryVmedisPayload]

	// VmedisID and VoucherNo are set when the entry was saved.
	VmedisID  *int
	VoucherNo *string
	// Error is the last failure.
	Error *string

	// CreatedByEmail is the login account that made the request (X-Email).
	CreatedByEmail string
}

// JournalEntryRequestPayload is a requested journal entry.
type JournalEntryRequestPayload struct {
	// Date is YYYY-MM-DD.
	Date              string `json:"date"`
	Description       string `json:"description"`
	CreditAccountCode string `json:"creditAccountCode"`
	DebitAccountCode  string `json:"debitAccountCode"`
	// Amount is whole Rupiah.
	Amount int64 `json:"amount"`
}

// JournalEntryVmedisPayload is what was sent to vmedis to save a journal
// entry.
type JournalEntryVmedisPayload struct {
	Date            string `json:"date"`
	VoucherNo       string `json:"voucherNo"`
	Description     string `json:"description"`
	DebitAccountID  int    `json:"debitAccountId"`
	CreditAccountID int    `json:"creditAccountId"`
	Amount          int64  `json:"amount"`
}
