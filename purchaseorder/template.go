package purchaseorder

import (
	"io"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// numberIssuerCode identifies the pharmacy in purchase order numbers.
const numberIssuerCode = "AAF"

// Template is a kind of purchase order: how it is numbered and rendered.
// Adding a template (e.g. Prekursor or OOT) means adding an entry to
// templates.
type Template struct {
	ID          string
	Name        string
	Description string
	// NumberCode goes into the number, e.g. "REG" in 001/SP-REG/AAF/XI/2026.
	NumberCode string
	// Render writes the purchase order as a PDF.
	Render func(w io.Writer, doc Document) error
}

// Info returns the template description shown to the dashboard.
func (t Template) Info() TemplateInfo {
	return TemplateInfo{ID: t.ID, Name: t.Name, Description: t.Description}
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
}

func documentFromDB(po models.PurchaseOrder) Document {
	return Document{
		Number:     po.Number,
		CreatedAt:  po.CreatedAt,
		Supplier:   po.SupplierSnapshot.Data(),
		Letterhead: po.LetterheadSnapshot.Data(),
		Signer:     po.SignerSnapshot.Data(),
		Items:      po.Items,
	}
}

var templates = []Template{
	{
		ID:          "reguler",
		Name:        "Surat Pesanan Reguler",
		Description: "Untuk obat bebas, bebas terbatas, keras dan alat kesehatan.",
		NumberCode:  "REG",
		Render:      renderReguler,
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
