package purchaseorder

import (
	"testing"
	"time"
)

func TestFormatNumber(t *testing.T) {
	if got := formatNumber("REG", 2026, 1, 7); got != "007/SP-REG/AAF/I/2026" {
		t.Errorf("got %q", got)
	}
	if got := formatNumber("REG", 2026, 12, 1234); got != "1234/SP-REG/AAF/XII/2026" {
		t.Errorf("got %q", got)
	}
}

func TestFormatLetterDate(t *testing.T) {
	// 4 March 20:00 UTC is 5 March in Jakarta.
	if got := formatLetterDate(time.Date(2026, time.March, 4, 20, 0, 0, 0, time.UTC)); got != "5 Maret 2026" {
		t.Errorf("got %q, want 5 Maret 2026", got)
	}
}

func TestPDFFilename(t *testing.T) {
	got := pdfFilename(
		time.Date(2026, time.November, 14, 20, 0, 0, 0, time.UTC),
		`PT. A/B: "Farma"  Ñusantara`,
		"Surat Pesanan Reguler",
		"001/SP-REG/AAF/XI/2026",
	)

	if want := "2026-11-15 - PT. AB Farma Ñusantara - Surat Pesanan Reguler - 001-SP-REG-AAF-XI-2026.pdf"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got := contentDisposition("a Ñ.pdf"); got != `attachment; filename="a _.pdf"; filename*=UTF-8''a%20%C3%91.pdf` {
		t.Errorf("content disposition = %q", got)
	}
}
