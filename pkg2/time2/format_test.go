package time2_test

import (
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"
)

// These assert the Indonesian month names specifically: formatting used to go
// through lctime's package-global locale, which is derived from the host's
// LC_TIME / LC_ALL / LANG, so the output depended on the environment and
// panicked outright on a locale lctime did not know.

func TestFormatDate(t *testing.T) {
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("load Asia/Jakarta: %s", err)
	}

	got := time2.FormatDate(time.Date(2026, time.August, 9, 14, 30, 5, 0, jakarta))
	if want := "09 Agustus 2026"; got != want {
		t.Errorf("FormatDate() = %q, want %q", got, want)
	}
}

func TestFormatDateTime(t *testing.T) {
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("load Asia/Jakarta: %s", err)
	}

	got := time2.FormatDateTime(time.Date(2026, time.August, 9, 14, 30, 5, 0, jakarta))
	if want := "09 Agustus 2026, 14:30:05 WIB"; got != want {
		t.Errorf("FormatDateTime() = %q, want %q", got, want)
	}
}

// Every month name must be present, so an empty or partially loaded locale
// fails here instead of panicking deep inside a handler.
func TestFormatDateAllMonths(t *testing.T) {
	// Spelled as lctime's id_ID locale spells them — note "Pebruari", the
	// older spelling, rather than today's more usual "Februari".
	want := []string{
		"Januari", "Pebruari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}

	for i, monthName := range want {
		month := time.Month(i + 1)

		got := time2.FormatDate(time.Date(2026, month, 1, 0, 0, 0, 0, time.UTC))
		if expected := "01 " + monthName + " 2026"; got != expected {
			t.Errorf("FormatDate(%s) = %q, want %q", month, got, expected)
		}
	}
}
