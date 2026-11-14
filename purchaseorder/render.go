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
	tablePadding    = 1.5
	tableNoWidth    = 12.0
	tableQtyWidth   = 38.0
)

// letter is a purchase order PDF being written.
type letter struct {
	pdf *fpdf.Fpdf
	doc Document
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
	pdf.SetTitle("Surat Pesanan "+doc.Number, true)
	pdf.SetAuthor(doc.Letterhead.Name, true)

	pdf.AddUTF8FontFromBytes(fontFamily, "", carlitoRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", carlitoBold)
	pdf.RegisterImageOptionsReader(logoName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(logoPNG))

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont(fontFamily, "", 9)
		pdf.SetTextColor(68, 68, 68)
		pdf.CellFormat(0, 5, fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.AddPage()

	return &letter{pdf: pdf, doc: doc}
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
// name, license number, address and contacts, then a double rule.
func (l *letter) writeLetterhead() {
	pdf := l.pdf
	lh := l.doc.Letterhead

	const (
		logoSize   = 26.0
		nameHeight = 7.0
		lineHeight = 5.0
	)

	top := pdf.GetY()
	pdf.ImageOptions(logoName, pageMarginSide, top, logoSize, logoSize, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	lines := []string{
		"Nomor Izin: " + lh.LicenseNumber,
		lh.Address,
		fmt.Sprintf("Email: %s, Telepon: %s", lh.Email, lh.Phone),
	}

	textX := pageMarginSide + logoSize + 5
	textW := l.contentWidth() - logoSize - 5
	textHeight := nameHeight + float64(len(lines))*lineHeight

	pdf.SetXY(textX, top+(logoSize-textHeight)/2)
	pdf.SetFont(fontFamily, "B", 14)
	pdf.CellFormat(textW, nameHeight, strings.ToUpper(lh.Name), "", 2, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", 11)
	for _, line := range lines {
		pdf.CellFormat(textW, lineHeight, line, "", 2, "L", false, 0, "")
	}

	ruleY := top + logoSize + 3
	right := pageMarginSide + l.contentWidth()

	pdf.SetLineWidth(0.6)
	pdf.Line(pageMarginSide, ruleY, right, ruleY)
	pdf.SetLineWidth(0.2)
	pdf.Line(pageMarginSide, ruleY+1, right, ruleY+1)

	pdf.SetXY(pageMarginSide, ruleY+1)
}

// writeTitle writes the centered title and the purchase order number.
func (l *letter) writeTitle(title string) {
	pdf := l.pdf

	pdf.Ln(8)
	pdf.SetFont(fontFamily, "BU", 14)
	pdf.CellFormat(0, 7, title, "", 1, "C", false, 0, "")
	pdf.SetFont(fontFamily, "", bodyFontSize)
	pdf.CellFormat(0, bodyLineHeight, "Nomor: "+l.doc.Number, "", 1, "C", false, 0, "")
}

// writeRecipient writes who the letter is addressed to.
func (l *letter) writeRecipient() {
	pdf := l.pdf
	supplier := l.doc.Supplier

	pdf.Ln(8)
	pdf.SetFont(fontFamily, "", bodyFontSize)
	pdf.CellFormat(0, bodyLineHeight, "Kepada Yth.", "", 1, "L", false, 0, "")
	pdf.SetFont(fontFamily, "B", bodyFontSize)
	pdf.MultiCell(0, bodyLineHeight, supplier.Name, "", "L", false)
	pdf.SetFont(fontFamily, "", bodyFontSize)
	pdf.MultiCell(0, bodyLineHeight, joinNonEmpty(", ", supplier.Address, supplier.City), "", "L", false)
}

// writeParagraph writes body text, wrapped to the content width.
func (l *letter) writeParagraph(text string) {
	l.pdf.SetFont(fontFamily, "", bodyFontSize)
	l.pdf.MultiCell(0, bodyLineHeight, text, "", "L", false)
}

type itemRow struct {
	no        string
	nameLines []string
	qtyLines  []string
}

func (r itemRow) height() float64 {
	return float64(max(len(r.nameLines), len(r.qtyLines), 1))*tableLineHeight + 2*tablePadding
}

// writeItems writes the item table. The header repeats on every page, and the
// last row moves to the next page together with the tail (e.g. the closing
// and signature) when they would not fit after it, so the tail is never left
// alone on a page.
func (l *letter) writeItems(tailHeight float64) {
	pdf := l.pdf
	nameWidth := l.contentWidth() - tableNoWidth - tableQtyWidth

	pdf.SetFont(fontFamily, "", tableFontSize)
	rows := make([]itemRow, len(l.doc.Items))
	for i, item := range l.doc.Items {
		rows[i] = itemRow{
			no:        fmt.Sprintf("%d", i+1),
			nameLines: wrapText(pdf, item.DrugName, nameWidth-2*tablePadding),
			qtyLines:  wrapText(pdf, item.Quantity, tableQtyWidth-2*tablePadding),
		}
	}

	pdf.Ln(2)
	l.writeTableHeader(nameWidth)

	for i, row := range rows {
		needed := row.height()
		if i == len(rows)-1 {
			needed += tailHeight
		}

		if pdf.GetY()+needed > l.pageBottom() {
			pdf.AddPage()
			l.writeTableHeader(nameWidth)
		}

		l.writeTableRow(row, nameWidth)
	}
}

func (l *letter) writeTableHeader(nameWidth float64) {
	pdf := l.pdf
	h := tableLineHeight + 2*tablePadding

	pdf.SetFont(fontFamily, "B", tableFontSize)
	pdf.SetLineWidth(0.2)
	pdf.SetX(pageMarginSide)
	pdf.CellFormat(tableNoWidth, h, "No.", "1", 0, "C", false, 0, "")
	pdf.SetCellMargin(tablePadding)
	pdf.CellFormat(nameWidth, h, "Nama Obat", "1", 0, "L", false, 0, "")
	pdf.SetCellMargin(0)
	pdf.CellFormat(tableQtyWidth, h, "Jumlah", "1", 1, "C", false, 0, "")
}

func (l *letter) writeTableRow(row itemRow, nameWidth float64) {
	pdf := l.pdf
	h := row.height()
	x, y := pageMarginSide, pdf.GetY()

	pdf.SetFont(fontFamily, "", tableFontSize)
	pdf.SetLineWidth(0.2)

	pdf.Rect(x, y, tableNoWidth, h, "D")
	pdf.Rect(x+tableNoWidth, y, nameWidth, h, "D")
	pdf.Rect(x+tableNoWidth+nameWidth, y, tableQtyWidth, h, "D")

	writeLines := func(lines []string, cellX, cellW float64, align string) {
		for i, line := range lines {
			pdf.SetXY(cellX+tablePadding, y+tablePadding+float64(i)*tableLineHeight)
			pdf.CellFormat(cellW-2*tablePadding, tableLineHeight, line, "", 0, align, false, 0, "")
		}
	}

	writeLines([]string{row.no}, x, tableNoWidth, "C")
	writeLines(row.nameLines, x+tableNoWidth, nameWidth, "L")
	writeLines(row.qtyLines, x+tableNoWidth+nameWidth, tableQtyWidth, "C")

	pdf.SetXY(pageMarginSide, y+h)
}

// signatureHeight is the height of the block written by writeSignature.
func signatureHeight(signed bool) float64 {
	if signed {
		return 10 + 2*bodyLineHeight + 22 + 2*bodyLineHeight
	}

	return 10 + 3*bodyLineHeight
}

// writeSignature writes the right-aligned place and date, then either the
// pharmacist's signature block or the pharmacy's name.
func (l *letter) writeSignature() {
	pdf := l.pdf
	const width = 75.0

	x := pageMarginSide + l.contentWidth() - width
	line := func(style, text string) {
		pdf.SetX(x)
		pdf.SetFont(fontFamily, style, bodyFontSize)
		pdf.CellFormat(width, bodyLineHeight, text, "", 1, "L", false, 0, "")
	}

	pdf.Ln(10)
	line("", fmt.Sprintf("%s, %s", l.doc.Letterhead.City, formatLetterDate(l.doc.CreatedAt)))

	if signer := l.doc.Signer; signer != nil {
		line("", "Apoteker Penanggung Jawab,")
		pdf.Ln(22)
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
		items[i] = item
	}
	doc.Items = items

	return doc
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
