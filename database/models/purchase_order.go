package models

import (
	"time"

	"gorm.io/datatypes"
)

// PurchaseOrder is a surat pesanan (SP): a numbered purchase order letter sent
// to a supplier. It keeps snapshots of everything printed on it, so the same
// PDF can be rendered again even after the supplier or the config changes.
// Purchase orders are never deleted by the app, and only edited to be
// cancelled.
type PurchaseOrder struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index;not null"`

	// TemplateCode, Year, Month and Sequence identify the purchase order in its
	// template's monthly numbering series. Year and Month are of CreatedAt in
	// Asia/Jakarta.
	TemplateCode string `gorm:"index;not null;uniqueIndex:idx_purchase_order_sequence,priority:1"`
	Year         int    `gorm:"not null;uniqueIndex:idx_purchase_order_sequence,priority:2"`
	Month        int    `gorm:"not null;uniqueIndex:idx_purchase_order_sequence,priority:3"`
	Sequence     int    `gorm:"not null;uniqueIndex:idx_purchase_order_sequence,priority:4"`
	// Number is the formatted number printed on the letter, e.g. 001/SP-REG/AAF/XI/2026.
	Number string `gorm:"uniqueIndex;not null"`

	SupplierVmedisID int `gorm:"index;not null"`
	SupplierSnapshot datatypes.JSONType[PurchaseOrderSupplier]

	IncludePharmacistSignature bool `gorm:"not null"`
	// SignerSnapshot is nil when the pharmacist signature is not included.
	SignerSnapshot     datatypes.JSONType[*PurchaseOrderSigner]
	LetterheadSnapshot datatypes.JSONType[PurchaseOrderLetterhead]

	// CreatedByEmployeeID is the HRIS employee the operator picked, since the
	// dashboard login is shared between staff.
	CreatedByEmployeeID   int    `gorm:"index;not null"`
	CreatedByEmployeeName string `gorm:"not null"`
	// CreatedByEmail is the login account that made the request, for audit.
	CreatedByEmail string

	// IdempotencyKey is chosen by the client once per create attempt, so a
	// retry after a lost response returns this purchase order instead of
	// creating another one. Nil for requests without a key.
	IdempotencyKey *string `gorm:"uniqueIndex"`

	// CancelledAt is set when the purchase order is cancelled (dibatalkan).
	// A cancelled purchase order keeps its number, which is never reused.
	// The other Cancelled fields are set with it.
	CancelledAt     *time.Time `gorm:"index"`
	CancelledReason *string
	// CancelledByEmployeeID is the HRIS employee the operator picked.
	CancelledByEmployeeID   *int
	CancelledByEmployeeName *string
	// CancelledByEmail is the login account that cancelled it, for audit.
	CancelledByEmail *string

	Items []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID"`
}

// PurchaseOrderSupplier is the snapshot of the supplier a purchase order is
// addressed to.
type PurchaseOrderSupplier struct {
	VmedisID int    `json:"vmedisId"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Phone    string `json:"phone"`
}

// PurchaseOrderSigner is the snapshot of the pharmacist who signs a purchase
// order (Apoteker Penanggung Jawab).
type PurchaseOrderSigner struct {
	Name string `json:"name"`
	// LicenseNumber is the pharmacist's SIPA.
	LicenseNumber string `json:"licenseNumber"`
}

// PurchaseOrderLetterhead is the snapshot of the pharmacy details printed on
// a purchase order.
type PurchaseOrderLetterhead struct {
	Name          string `json:"name"`
	LicenseNumber string `json:"licenseNumber"`
	Address       string `json:"address"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	// City is where the letter is written from, printed before its date.
	City string `json:"city"`
}

// PurchaseOrderItem is a drug ordered in a purchase order.
type PurchaseOrderItem struct {
	ID              uint `gorm:"primarykey"`
	PurchaseOrderID uint `gorm:"index;not null"`
	// Position is the 1-based order of the item on the letter.
	Position int `gorm:"not null"`
	// DrugVmedisCode is nil for drugs added by hand that are not in vmedis.
	DrugVmedisCode *string
	DrugName       string `gorm:"index;not null"`
	// Quantity is the quantity as typed, e.g. "10 Box". On templates with
	// controlled item fields it is a whole number, e.g. "12".
	Quantity string `gorm:"not null"`

	// The fields below are set only on templates with controlled item fields
	// (Prekursor and OOT), nil otherwise.

	// ActiveIngredient (zat aktif) and DosageFormStrength (bentuk dan
	// kekuatan sediaan) are snapshots of the drug's data in vmedis.
	ActiveIngredient   *string
	DosageFormStrength *string
	// SmallestPackage is the unit the quantity is in (kemasan terkecil),
	// e.g. "Box @ 10 Strip".
	SmallestPackage *string
	// QuantityWords is the quantity in Indonesian words, e.g. "dua belas".
	QuantityWords *string
	// Note is the item's keterangan, nil when empty.
	Note *string
}

// PurchaseOrderSequence holds the last number used in a template's monthly
// numbering series.
type PurchaseOrderSequence struct {
	TemplateCode string `gorm:"primaryKey"`
	Year         int    `gorm:"primaryKey;autoIncrement:false"`
	Month        int    `gorm:"primaryKey;autoIncrement:false"`
	LastSequence int    `gorm:"not null"`
}
