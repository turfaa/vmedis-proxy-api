package purchaseorder

import (
	"io"
	"slices"
	"strings"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// controlledText is the wording that differs between the controlled
// templates.
type controlledText struct {
	title string
	// orderText introduces the supplier, e.g. "Mengajukan pesanan Obat-Obat
	// Tertentu kepada:".
	orderText string
	// activeIngredientHeader is the header of the active ingredient column.
	activeIngredientHeader string
	// purposeText introduces the pharmacy the drugs are for.
	purposeText string
}

// renderPrekursor renders a Surat Pesanan Prekursor Farmasi: the order
// letter for drugs containing prekursor farmasi.
func renderPrekursor(w io.Writer, doc Document) error {
	return renderControlled(w, doc, controlledText{
		title:                  "SURAT PESANAN OBAT MENGANDUNG PREKURSOR FARMASI",
		orderText:              "Mengajukan pesanan obat mengandung prekursor farmasi kepada:",
		activeIngredientHeader: "Zat Aktif Prekursor Farmasi",
		purposeText:            "Obat mengandung prekursor farmasi tersebut akan dipergunakan untuk memenuhi kebutuhan:",
	})
}

// renderOOT renders a Surat Pesanan Obat-Obat Tertentu: the order letter
// for drugs often misused (obat-obat tertentu).
func renderOOT(w io.Writer, doc Document) error {
	return renderControlled(w, doc, controlledText{
		title:                  "SURAT PESANAN OBAT-OBAT TERTENTU",
		orderText:              "Mengajukan pesanan Obat-Obat Tertentu kepada:",
		activeIngredientHeader: "Zat Aktif Obat-Obat Tertentu",
		purposeText:            "Obat-Obat Tertentu tersebut akan dipergunakan untuk memenuhi kebutuhan:",
	})
}

// renderControlled renders the order letter for controlled drugs, with the
// contents BPOM requires: the pharmacist ordering, the supplier, each drug's
// active ingredient, dosage form and strength, smallest package and quantity
// in numbers and words, the pharmacy the drugs are for, and the
// pharmacist's name and SIPA with space for a wet signature and stamp.
func renderControlled(w io.Writer, doc Document, text controlledText) error {
	l := newLetter(doc)
	l.bodyFontSize = controlledBodyFontSize
	l.bodyLineHeight = controlledBodyLineHeight

	// Controlled templates always store a signer; the zero value only keeps
	// a broken purchase order printable.
	if l.doc.Signer == nil {
		l.doc.Signer = &models.PurchaseOrderSigner{}
	}
	signer := l.doc.Signer
	supplier, letterhead := l.doc.Supplier, l.doc.Letterhead

	l.writeLetterhead()
	l.writeTitle(text.title)

	l.pdf.Ln(4)
	l.writeParagraph("Yang bertanda tangan di bawah ini:")
	l.writeFields([]field{
		{label: "Nama", value: "apt. " + signer.Name},
		{label: "Jabatan", value: "Apoteker Penanggung Jawab"},
		{label: "Nomor SIPA", value: signer.LicenseNumber},
	})

	l.pdf.Ln(2)
	l.writeParagraph(text.orderText)
	l.writeFields([]field{
		{label: "Nama PBF", value: supplier.Name, bold: true},
		{label: "Alamat", value: joinNonEmpty(", ", supplier.Address, supplier.City)},
		{label: "Telepon", value: supplier.Phone},
	})

	l.pdf.Ln(2)
	l.writeParagraph("dengan obat yang dipesan sebagai berikut:")

	purpose := []field{
		{label: "Nama Apotek", value: letterhead.Name},
		{label: "Alamat", value: letterhead.Address},
		{label: "Surat Izin Apotek", value: letterhead.LicenseNumber},
	}

	const purposeGap = 3.0
	tail := purposeGap + l.paragraphHeight(text.purposeText) + l.fieldsHeight(purpose) + l.signatureHeight()
	l.writeItems(controlledTable(l.doc.Items, text.activeIngredientHeader), tail)

	l.pdf.Ln(purposeGap)
	l.writeParagraph(text.purposeText)
	l.writeFields(purpose)
	l.writeSignature()

	return l.output(w)
}

// controlledTable is the item table of the controlled templates. The Ket.
// column is left out when no item has a note.
func controlledTable(items []models.PurchaseOrderItem, activeIngredientHeader string) itemTable {
	text := func(get func(models.PurchaseOrderItem) *string) func(int, models.PurchaseOrderItem) string {
		return func(_ int, item models.PurchaseOrderItem) string { return deref(get(item)) }
	}

	columns := []tableColumn{
		{header: "No.", width: 8, align: "C", value: itemNumber},
		{header: "Nama Obat", weight: 1.3, align: "L", value: func(_ int, item models.PurchaseOrderItem) string { return item.DrugName }},
		{header: activeIngredientHeader, weight: 1.2, align: "L", value: text(func(item models.PurchaseOrderItem) *string { return item.ActiveIngredient })},
		{header: "Bentuk dan Kekuatan Sediaan", weight: 1.1, align: "L", value: text(func(item models.PurchaseOrderItem) *string { return item.DosageFormStrength })},
		{header: "Satuan (Kemasan Terkecil)", weight: 1, align: "L", value: text(func(item models.PurchaseOrderItem) *string { return item.SmallestPackage })},
		{header: "Jumlah (Angka)", width: 16, align: "C", value: func(_ int, item models.PurchaseOrderItem) string { return item.Quantity }},
		{header: "Jumlah (Huruf)", weight: 1.1, align: "L", value: text(func(item models.PurchaseOrderItem) *string { return item.QuantityWords })},
	}

	hasNote := slices.ContainsFunc(items, func(item models.PurchaseOrderItem) bool {
		return strings.TrimSpace(deref(item.Note)) != ""
	})
	if hasNote {
		columns = append(columns, tableColumn{
			header: "Ket.", weight: 0.8, align: "L",
			value: text(func(item models.PurchaseOrderItem) *string { return item.Note }),
		})
	}

	return itemTable{
		fontSize:   controlledTableFontSize,
		lineHeight: controlledTableLineHeight,
		columns:    columns,
	}
}
