package purchaseorder

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

var (
	//go:embed assets/Carlito-Regular.ttf
	carlitoRegular []byte
	//go:embed assets/Carlito-Bold.ttf
	carlitoBold []byte
	//go:embed assets/logo.png
	logoPNG []byte
)

const (
	fontFamily = "Carlito"
	logoName   = "logo"

	pageMarginTop    = 15.0
	pageMarginSide   = 20.0
	pageMarginBottom = 20.0

	bodyFontSize   = 12.0
	bodyLineHeight = 6.0

	tableFontSize   = 11.0
	tableLineHeight = 5.0
	tablePadding    = 1.0
	tableNoWidth    = 12.0
	tableQtyWidth   = 38.0

	// Controlled templates (Prekursor and OOT) have more text and columns,
	// so they use smaller type.
	controlledBodyFontSize    = 11.0
	controlledBodyLineHeight  = 5.5
	controlledTableFontSize   = 10.0
	controlledTableLineHeight = 4.5

	// fieldLabelWidth is the width of the labels of writeFields, e.g. "Nama",
	// and fieldIndent the indent of their block.
	fieldLabelWidth = 35.0
	fieldColonWidth = 4.0
	fieldIndent     = 4.0

	watermarkText     = "DIBATALKAN"
	watermarkFontSize = 100.0

	// signatureGap is the space above the place and date, and signatureSpace
	// the space left for the pharmacist to sign in.
	signatureGap   = 8.0
	signatureSpace = 18.0
)

// letter is a purchase order PDF being written.
type letter struct {
	pdf *fpdf.Fpdf
	doc Document

	// bodyFontSize and bodyLineHeight are of the letter's text outside the
	// item table.
	bodyFontSize   float64
	bodyLineHeight float64
}

func newLetter(doc Document) *letter {
	doc = pdfSafeDocument(doc)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pageMarginSide, pageMarginTop, pageMarginSide)
	pdf.SetAutoPageBreak(true, pageMarginBottom)
	pdf.SetCellMargin(0)
	pdf.AliasNbPages("{nb}")

	// Fixed dates keep the output the same every time a stored purchase order
	// is downloaded again.
	pdf.SetCreationDate(doc.CreatedAt)
	pdf.SetModificationDate(doc.CreatedAt)
	// fpdf otherwise writes fonts and images in map order, which changes
	// from run to run.
	pdf.SetCatalogSort(true)
	pdf.SetTitle("Surat Pesanan "+doc.Number, true)
	pdf.SetAuthor(doc.Letterhead.Name, true)

	pdf.AddUTF8FontFromBytes(fontFamily, "", carlitoRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", carlitoBold)
	pdf.RegisterImageOptionsReader(logoName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(logoPNG))

	pdf.SetFooterFunc(func() {
		if doc.Cancellation != nil {
			writeWatermark(pdf)
		}

		pdf.SetY(-12)
		pdf.SetFont(fontFamily, "", 9)
		pdf.SetTextColor(68, 68, 68)
		pdf.CellFormat(0, 5, fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.AddPage()

	return &letter{pdf: pdf, doc: doc, bodyFontSize: bodyFontSize, bodyLineHeight: bodyLineHeight}
}

// writeWatermark writes a light, diagonal DIBATALKAN across the page.
func writeWatermark(pdf *fpdf.Fpdf) {
	w, h := pdf.GetPageSize()
	cx, cy := w/2, h/2

	pdf.SetFont(fontFamily, "B", watermarkFontSize)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetAlpha(0.12, "Normal")
	pdf.TransformBegin()
	pdf.TransformRotate(35, cx, cy)

	textWidth := pdf.GetStringWidth(watermarkText)
	_, fontHeight := pdf.GetFontSize()
	pdf.Text(cx-textWidth/2, cy+fontHeight/3, watermarkText)

	pdf.TransformEnd()
	pdf.SetAlpha(1, "Normal")
}

func (l *letter) contentWidth() float64 {
	w, _ := l.pdf.GetPageSize()
	return w - 2*pageMarginSide
}

func (l *letter) pageBottom() float64 {
	_, h := l.pdf.GetPageSize()
	return h - pageMarginBottom
}

func (l *letter) output(w io.Writer) error {
	if err := l.pdf.Output(w); err != nil {
		return fmt.Errorf("write pdf: %w", err)
	}

	return nil
}

// writeLetterhead writes the pharmacy letterhead: the logo, the pharmacy's
// name, license number, address and contacts, then a double rule. It is kept
// compact so the page is left for the items.
func (l *letter) writeLetterhead() {
	pdf := l.pdf
	lh := l.doc.Letterhead

	const (
		logoHeight     = 14.0
		logoGap        = 5.0
		nameFontSize   = 12.0
		nameHeight     = 5.5
		detailFontSize = 9.5
		lineHeight     = 4.0
	)

	info := pdf.GetImageInfo(logoName)
	logoWidth := logoHeight * info.Width() / info.Height()

	textX := pageMarginSide + logoWidth + logoGap
	textW := l.contentWidth() - logoWidth - logoGap

	pdf.SetFont(fontFamily, "", detailFontSize)
	var lines []string
	lines = append(lines, "Nomor Izin: "+lh.LicenseNumber)
	lines = append(lines, wrapText(pdf, lh.Address, textW)...)
	lines = append(lines, fmt.Sprintf("Email: %s, Telepon: %s", lh.Email, lh.Phone))

	textHeight := nameHeight + float64(len(lines))*lineHeight
	blockHeight := max(logoHeight, textHeight)

	top := pdf.GetY()
	pdf.ImageOptions(logoName, pageMarginSide, top+(blockHeight-logoHeight)/2, logoWidth, logoHeight, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	pdf.SetXY(textX, top+(blockHeight-textHeight)/2)
	pdf.SetFont(fontFamily, "B", nameFontSize)
	pdf.CellFormat(textW, nameHeight, strings.ToUpper(lh.Name), "", 2, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", detailFontSize)
	for _, line := range lines {
		pdf.CellFormat(textW, lineHeight, line, "", 2, "L", false, 0, "")
	}

	ruleY := top + blockHeight + 2
	right := pageMarginSide + l.contentWidth()

	pdf.SetLineWidth(0.6)
	pdf.Line(pageMarginSide, ruleY, right, ruleY)
	pdf.SetLineWidth(0.2)
	pdf.Line(pageMarginSide, ruleY+1, right, ruleY+1)

	pdf.SetXY(pageMarginSide, ruleY+1)
}

// writeTitle writes the centered title and the purchase order number, and
// under them when the purchase order is cancelled, when, by whom and why.
func (l *letter) writeTitle(title string) {
	pdf := l.pdf

	pdf.Ln(4)
	pdf.SetFont(fontFamily, "BU", 14)
	pdf.MultiCell(0, 7, title, "", "C", false)
	pdf.SetFont(fontFamily, "", l.bodyFontSize)
	pdf.CellFormat(0, l.bodyLineHeight, "Nomor: "+l.doc.Number, "", 1, "C", false, 0, "")

	if c := l.doc.Cancellation; c != nil {
		pdf.Ln(2)
		pdf.SetFont(fontFamily, "B", l.bodyFontSize-1)
		pdf.SetLineWidth(0.4)
		pdf.SetCellMargin(2)
		pdf.MultiCell(0, l.bodyLineHeight, fmt.Sprintf(
			"DIBATALKAN pada %s oleh %s. Alasan: %s",
			formatLetterDate(c.At), c.EmployeeName, c.Reason,
		), "1", "L", false)
		pdf.SetCellMargin(0)
		pdf.SetLineWidth(0.2)
	}
}

// writeRecipient writes who the letter is addressed to.
func (l *letter) writeRecipient() {
	pdf := l.pdf
	supplier := l.doc.Supplier

	pdf.Ln(5)
	pdf.SetFont(fontFamily, "", l.bodyFontSize)
	pdf.CellFormat(0, l.bodyLineHeight, "Kepada Yth.", "", 1, "L", false, 0, "")
	pdf.SetFont(fontFamily, "B", l.bodyFontSize)
	pdf.MultiCell(0, l.bodyLineHeight, supplier.Name, "", "L", false)
	pdf.SetFont(fontFamily, "", l.bodyFontSize)
	pdf.MultiCell(0, l.bodyLineHeight, joinNonEmpty(", ", supplier.Address, supplier.City), "", "L", false)
}

// writeParagraph writes body text, wrapped to the content width.
func (l *letter) writeParagraph(text string) {
	l.pdf.SetFont(fontFamily, "", l.bodyFontSize)
	l.pdf.MultiCell(0, l.bodyLineHeight, text, "", "L", false)
}

// paragraphHeight is the height writeParagraph writes text in.
func (l *letter) paragraphHeight(text string) float64 {
	l.pdf.SetFont(fontFamily, "", l.bodyFontSize)
	return float64(len(wrapText(l.pdf, text, l.contentWidth()))) * l.bodyLineHeight
}

// field is a labelled value, e.g. "Nama : apt. Nama".
type field struct {
	label string
	value string
	bold  bool
}

func (l *letter) fieldValueWidth() float64 {
	return l.contentWidth() - fieldIndent - fieldLabelWidth - fieldColonWidth
}

func (l *letter) fieldLines(f field) []string {
	style := ""
	if f.bold {
		style = "B"
	}

	l.pdf.SetFont(fontFamily, style, l.bodyFontSize)
	return wrapText(l.pdf, f.value, l.fieldValueWidth())
}

// writeFields writes indented labelled values, with the labels and colons
// lined up.
func (l *letter) writeFields(fields []field) {
	pdf := l.pdf
	labelX := pageMarginSide + fieldIndent
	valueX := labelX + fieldLabelWidth + fieldColonWidth

	for _, f := range fields {
		lines := l.fieldLines(f)

		pdf.SetFont(fontFamily, "", l.bodyFontSize)
		pdf.SetX(labelX)
		pdf.CellFormat(fieldLabelWidth, l.bodyLineHeight, f.label, "", 0, "L", false, 0, "")
		pdf.CellFormat(fieldColonWidth, l.bodyLineHeight, ":", "", 0, "L", false, 0, "")

		if f.bold {
			pdf.SetFont(fontFamily, "B", l.bodyFontSize)
		}
		for i, line := range lines {
			if i > 0 {
				pdf.SetX(valueX)
			}
			pdf.CellFormat(l.fieldValueWidth(), l.bodyLineHeight, line, "", 1, "L", false, 0, "")
		}
	}
}

// fieldsHeight is the height writeFields writes fields in.
func (l *letter) fieldsHeight(fields []field) float64 {
	height := 0.0
	for _, f := range fields {
		height += float64(len(l.fieldLines(f))) * l.bodyLineHeight
	}

	return height
}

// tableColumn is a column of the item table.
type tableColumn struct {
	header string
	// width is the column's width in mm. Columns with no width share the
	// width left by their weight.
	width  float64
	weight float64
	// align is the alignment of the header and the cells: "L", "C" or "R".
	align string
	// value is the cell of the item at the 0-based index.
	value func(i int, item models.PurchaseOrderItem) string
}

// itemTable is how the items are laid out.
type itemTable struct {
	fontSize   float64
	lineHeight float64
	columns    []tableColumn
}

// tableRow is a row of the table, its cells wrapped to their columns.
type tableRow struct {
	cells      [][]string
	lineHeight float64
}

func (r tableRow) height() float64 {
	lines := 1
	for _, cell := range r.cells {
		lines = max(lines, len(cell))
	}

	return float64(lines)*r.lineHeight + 2*tablePadding
}

// columnWidths returns the width of each column: fixed widths as they are,
// and the rest of the content width shared by weight.
func (l *letter) columnWidths(t itemTable) []float64 {
	fixed, weights := 0.0, 0.0
	for _, c := range t.columns {
		if c.width > 0 {
			fixed += c.width
		} else {
			weights += c.weight
		}
	}

	left := l.contentWidth() - fixed
	widths := make([]float64, len(t.columns))
	for i, c := range t.columns {
		if c.width > 0 {
			widths[i] = c.width
		} else if weights > 0 {
			widths[i] = left * c.weight / weights
		}
	}

	return widths
}

func (l *letter) tableRow(t itemTable, widths []float64, style string, cell func(i int, c tableColumn) string) tableRow {
	l.pdf.SetFont(fontFamily, style, t.fontSize)

	row := tableRow{cells: make([][]string, len(t.columns)), lineHeight: t.lineHeight}
	for i, c := range t.columns {
		row.cells[i] = wrapText(l.pdf, cell(i, c), widths[i]-2*tablePadding)
	}

	return row
}

// writeItems writes the item table. The header repeats on every page, and the
// last row moves to the next page together with the tail (e.g. the closing
// and signature) when they would not fit after it, so the tail is never left
// alone on a page.
func (l *letter) writeItems(t itemTable, tailHeight float64) {
	pdf := l.pdf
	widths := l.columnWidths(t)

	header := l.tableRow(t, widths, "B", func(_ int, c tableColumn) string { return c.header })
	rows := make([]tableRow, len(l.doc.Items))
	for i, item := range l.doc.Items {
		rows[i] = l.tableRow(t, widths, "", func(_ int, c tableColumn) string { return c.value(i, item) })
	}

	pdf.Ln(2)
	l.writeTableRow(t, widths, header, "B")

	for i, row := range rows {
		needed := row.height()
		if i == len(rows)-1 {
			needed += tailHeight
		}

		if pdf.GetY()+needed > l.pageBottom() {
			pdf.AddPage()
			l.writeTableRow(t, widths, header, "B")
		}

		l.writeTableRow(t, widths, row, "")
	}
}

func (l *letter) writeTableRow(t itemTable, widths []float64, row tableRow, style string) {
	pdf := l.pdf
	h := row.height()
	y := pdf.GetY()

	pdf.SetFont(fontFamily, style, t.fontSize)
	pdf.SetLineWidth(0.2)

	x := pageMarginSide
	for i, c := range t.columns {
		pdf.Rect(x, y, widths[i], h, "D")
		for j, line := range row.cells[i] {
			pdf.SetXY(x+tablePadding, y+tablePadding+float64(j)*t.lineHeight)
			pdf.CellFormat(widths[i]-2*tablePadding, t.lineHeight, line, "", 0, c.align, false, 0, "")
		}
		x += widths[i]
	}

	pdf.SetXY(pageMarginSide, y+h)
}

func itemNumber(i int, _ models.PurchaseOrderItem) string {
	return fmt.Sprintf("%d", i+1)
}

// regulerTable is the table of a Surat Pesanan Reguler: the drug's name and
// the quantity as typed.
var regulerTable = itemTable{
	fontSize:   tableFontSize,
	lineHeight: tableLineHeight,
	columns: []tableColumn{
		{header: "No.", width: tableNoWidth, align: "C", value: itemNumber},
		{header: "Nama Obat", weight: 1, align: "L", value: func(_ int, item models.PurchaseOrderItem) string { return item.DrugName }},
		{header: "Jumlah", width: tableQtyWidth, align: "C", value: func(_ int, item models.PurchaseOrderItem) string { return item.Quantity }},
	},
}

// signatureHeight is the height of the block written by writeSignature.
func (l *letter) signatureHeight() float64 {
	if l.doc.Signer != nil {
		return signatureGap + 2*l.bodyLineHeight + signatureSpace + 2*l.bodyLineHeight
	}

	return signatureGap + 3*l.bodyLineHeight
}

// writeSignature writes the right-aligned place and date, then either the
// pharmacist's signature block or the pharmacy's name.
func (l *letter) writeSignature() {
	pdf := l.pdf
	const width = 75.0

	x := pageMarginSide + l.contentWidth() - width
	line := func(style, text string) {
		pdf.SetX(x)
		pdf.SetFont(fontFamily, style, l.bodyFontSize)
		pdf.CellFormat(width, l.bodyLineHeight, text, "", 1, "L", false, 0, "")
	}

	pdf.Ln(signatureGap)
	line("", fmt.Sprintf("%s, %s", l.doc.Letterhead.City, formatLetterDate(l.doc.CreatedAt)))

	if signer := l.doc.Signer; signer != nil {
		line("", "Apoteker Penanggung Jawab,")
		pdf.Ln(signatureSpace)
		line("BU", "apt. "+signer.Name)
		line("", "SIPA: "+signer.LicenseNumber)
		return
	}

	line("", "Hormat kami,")
	line("B", l.doc.Letterhead.Name)
}

// wrapText splits text into lines no wider than width in the current font,
// breaking at spaces, and inside words only when a word is too long.
func wrapText(pdf *fpdf.Fpdf, text string, width float64) []string {
	var lines []string

	for _, paragraph := range strings.Split(text, "\n") {
		current := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := word
			if current != "" {
				candidate = current + " " + word
			}

			if pdf.GetStringWidth(candidate) <= width {
				current = candidate
				continue
			}

			if current != "" {
				lines = append(lines, current)
				current = ""
			}

			for pdf.GetStringWidth(word) > width {
				cut := fittingPrefix(pdf, word, width)
				lines = append(lines, word[:cut])
				word = word[cut:]
			}
			current = word
		}

		if current != "" {
			lines = append(lines, current)
		}
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// fittingPrefix returns the byte length of the longest prefix of word that
// fits in width, but at least one rune.
func fittingPrefix(pdf *fpdf.Fpdf, word string, width float64) int {
	cut := 0
	for i, r := range word {
		end := i + utf8.RuneLen(r)
		if cut > 0 && pdf.GetStringWidth(word[:end]) > width {
			break
		}
		cut = end
	}

	return cut
}

func joinNonEmpty(sep string, parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}

	return strings.Join(nonEmpty, sep)
}

// pdfSafeDocument returns the document with every printed text made safe for
// fpdf.
func pdfSafeDocument(doc Document) Document {
	doc.Number = pdfSafe(doc.Number)

	doc.Supplier.Name = pdfSafe(doc.Supplier.Name)
	doc.Supplier.Address = pdfSafe(doc.Supplier.Address)
	doc.Supplier.City = pdfSafe(doc.Supplier.City)

	doc.Letterhead.Name = pdfSafe(doc.Letterhead.Name)
	doc.Letterhead.LicenseNumber = pdfSafe(doc.Letterhead.LicenseNumber)
	doc.Letterhead.Address = pdfSafe(doc.Letterhead.Address)
	doc.Letterhead.Email = pdfSafe(doc.Letterhead.Email)
	doc.Letterhead.Phone = pdfSafe(doc.Letterhead.Phone)
	doc.Letterhead.City = pdfSafe(doc.Letterhead.City)

	if doc.Signer != nil {
		signer := *doc.Signer
		signer.Name = pdfSafe(signer.Name)
		signer.LicenseNumber = pdfSafe(signer.LicenseNumber)
		doc.Signer = &signer
	}

	items := make([]models.PurchaseOrderItem, len(doc.Items))
	for i, item := range doc.Items {
		item.DrugName = pdfSafe(item.DrugName)
		item.Quantity = pdfSafe(item.Quantity)
		item.ActiveIngredient = pdfSafePtr(item.ActiveIngredient)
		item.DosageFormStrength = pdfSafePtr(item.DosageFormStrength)
		item.SmallestPackage = pdfSafePtr(item.SmallestPackage)
		item.QuantityWords = pdfSafePtr(item.QuantityWords)
		item.Note = pdfSafePtr(item.Note)
		items[i] = item
	}
	doc.Items = items

	if doc.Cancellation != nil {
		c := *doc.Cancellation
		c.EmployeeName = pdfSafe(c.EmployeeName)
		c.Reason = pdfSafe(c.Reason)
		doc.Cancellation = &c
	}

	return doc
}

func pdfSafePtr(s *string) *string {
	if s == nil {
		return nil
	}

	safe := pdfSafe(*s)
	return &safe
}

// pdfSafe removes the characters fpdf cannot write: those outside the Basic
// Multilingual Plane (e.g. emoji), which make it panic, and control
// characters other than newlines.
func pdfSafe(s string) string {
	return strings.Map(func(r rune) rune {
		if r > 0xFFFF || r == utf8.RuneError || (unicode.IsControl(r) && r != '\n') {
			return -1
		}
		return r
	}, s)
}
