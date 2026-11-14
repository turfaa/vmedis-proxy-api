package purchaseorder

import (
	"fmt"
	"strings"
	"time"
)

var romanMonths = [...]string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}

var indonesianMonths = [...]string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

// formatNumber formats a purchase order number, e.g. 001/SP-REG/AAF/XI/2026.
func formatNumber(numberCode string, year, month, sequence int) string {
	return fmt.Sprintf("%03d/SP-%s/%s/%s/%d", sequence, numberCode, numberIssuerCode, romanMonths[month-1], year)
}

// formatLetterDate formats a date the way it is written on a letter, e.g.
// "5 November 2026".
func formatLetterDate(t time.Time) string {
	t = t.In(time.Local)
	return fmt.Sprintf("%d %s %d", t.Day(), indonesianMonths[t.Month()-1], t.Year())
}

// pdfFilename returns the name a purchase order PDF is downloaded as:
// "[YYYY-MM-DD] - [supplier] - [template] - [number].pdf", with the number's
// slashes turned into dashes and characters not allowed in filenames removed.
func pdfFilename(createdAt time.Time, supplierName, templateName, number string) string {
	return fmt.Sprintf(
		"%s - %s - %s - %s.pdf",
		createdAt.In(time.Local).Format(time.DateOnly),
		sanitizeFilenamePart(supplierName),
		sanitizeFilenamePart(templateName),
		sanitizeFilenamePart(strings.ReplaceAll(number, "/", "-")),
	)
}

func sanitizeFilenamePart(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || strings.ContainsRune(`\/:*?"<>|`, r) {
			return -1
		}
		return r
	}, s)

	return strings.Join(strings.Fields(s), " ")
}

// asciiFilename replaces the non-ASCII characters of a filename, for the
// plain filename parameter of Content-Disposition.
func asciiFilename(s string) string {
	return strings.Map(func(r rune) rune {
		if r > 0x7e {
			return '_'
		}
		return r
	}, s)
}
