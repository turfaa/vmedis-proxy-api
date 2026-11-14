package models

import (
	"time"

	"gorm.io/datatypes"
)

// PurchaseOrder is a surat pesanan (SP): a numbered purchase order letter sent
// to a supplier. It keeps snapshots of everything printed on it, so the same
// PDF can be rendered again even after the supplier or the config changes.
// Purchase orders are never edited or deleted by the app.
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
	// Quantity is the quantity as typed, e.g. "10 Box".
	Quantity string `gorm:"not null"`
	// Extra holds template-specific fields of the item.
	Extra datatypes.JSON
}

// PurchaseOrderSequence holds the last number used in a template's monthly
// numbering series.
type PurchaseOrderSequence struct {
	TemplateCode string `gorm:"primaryKey"`
	Year         int    `gorm:"primaryKey;autoIncrement:false"`
	Month        int    `gorm:"primaryKey;autoIncrement:false"`
	LastSequence int    `gorm:"not null"`
}
