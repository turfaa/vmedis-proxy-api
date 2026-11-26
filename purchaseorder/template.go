package purchaseorder

import (
	"io"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// numberIssuerCode identifies the pharmacy in purchase order numbers.
const numberIssuerCode = "AAF"

// ItemFields is what is filled in for each item of a purchase order.
type ItemFields string

const (
	// ItemFieldsBasic is a drug name and a free-text quantity, e.g. "10 Box".
	ItemFieldsBasic ItemFields = "basic"
	// ItemFieldsControlled is what BPOM requires on an order of controlled
	// drugs (prekursor farmasi, obat-obat tertentu): the drug's active
	// ingredient and dosage form and strength from vmedis, the smallest
	// package, a whole-number quantity, the quantity in words and an
	// optional note.
	ItemFieldsControlled ItemFields = "controlled"
)

// SignatureMode is how the pharmacist signs a purchase order.
type SignatureMode string

const (
	// SignatureModeOptional prints the pharmacist's signature block only
	// when asked; otherwise the pharmacy's name.
	SignatureModeOptional SignatureMode = "optional"
	// SignatureModeWet always prints the pharmacist's name and SIPA, with
	// empty space for a wet signature and stamp.
	SignatureModeWet SignatureMode = "wet"
)

// Template is a kind of purchase order: how it is numbered and rendered.
// Adding a template means adding an entry to templates.
type Template struct {
	ID          string
	Name        string
	Description string
	// NumberCode goes into the number, e.g. "REG" in 001/SP-REG/AAF/XI/2026.
	NumberCode    string
	ItemFields    ItemFields
	SignatureMode SignatureMode
	// DrugClass is the drug class the template is for, matched against the
	// drug's golongan in vmedis to warn about drugs of another class. Empty
	// for templates for any drug.
	DrugClass drugClass
	// Render writes the purchase order as a PDF.
	Render func(w io.Writer, doc Document) error
}

// Info returns the template description shown to the dashboard.
func (t Template) Info() TemplateInfo {
	return TemplateInfo{
		ID:            t.ID,
		Name:          t.Name,
		Description:   t.Description,
		ItemFields:    t.ItemFields,
		SignatureMode: t.SignatureMode,
	}
}

// Document is everything printed on a purchase order, taken from its stored
// snapshots.
type Document struct {
	Number     string
	CreatedAt  time.Time
	Supplier   models.PurchaseOrderSupplier
	Letterhead models.PurchaseOrderLetterhead
	// Signer is nil when the pharmacist signature is not included.
	Signer *models.PurchaseOrderSigner
	Items  []models.PurchaseOrderItem
	// Cancellation is nil when the purchase order is not cancelled.
	Cancellation *DocumentCancellation
}

// DocumentCancellation is printed on a cancelled purchase order.
type DocumentCancellation struct {
	At           time.Time
	EmployeeName string
	Reason       string
}

func documentFromDB(po models.PurchaseOrder) Document {
	doc := Document{
		Number:     po.Number,
		CreatedAt:  po.CreatedAt,
		Supplier:   po.SupplierSnapshot.Data(),
		Letterhead: po.LetterheadSnapshot.Data(),
		Signer:     po.SignerSnapshot.Data(),
		Items:      po.Items,
	}

	if po.CancelledAt != nil {
		doc.Cancellation = &DocumentCancellation{
			At:           *po.CancelledAt,
			EmployeeName: deref(po.CancelledByEmployeeName),
			Reason:       deref(po.CancelledReason),
		}
	}

	return doc
}

var templates = []Template{
	{
		ID:            "reguler",
		Name:          "Surat Pesanan Reguler",
		Description:   "Untuk obat bebas, bebas terbatas, keras dan alat kesehatan.",
		NumberCode:    "REG",
		ItemFields:    ItemFieldsBasic,
		SignatureMode: SignatureModeOptional,
		Render:        renderReguler,
	},
	{
		ID:            "prekursor",
		Name:          "Surat Pesanan Prekursor Farmasi",
		Description:   "Untuk obat mengandung prekursor farmasi. Dibuat terpisah dari obat lain.",
		NumberCode:    "PRE",
		ItemFields:    ItemFieldsControlled,
		SignatureMode: SignatureModeWet,
		DrugClass:     drugClassPrekursor,
		Render:        renderPrekursor,
	},
	{
		ID:            "oot",
		Name:          "Surat Pesanan Obat-Obat Tertentu",
		Description:   "Untuk Obat-Obat Tertentu yang sering disalahgunakan. Dibuat terpisah dari obat lain.",
		NumberCode:    "OOT",
		ItemFields:    ItemFieldsControlled,
		SignatureMode: SignatureModeWet,
		DrugClass:     drugClassOOT,
		Render:        renderOOT,
	},
}

func templateByID(id string) (Template, bool) {
	for _, t := range templates {
		if t.ID == id {
			return t, true
		}
	}

	return Template{}, false
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}

	return *p
}
