package purchaseorder

import "strings"

var terbilangDigits = [...]string{
	"nol", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan",
	"sepuluh", "sebelas",
}

// terbilangScales are the Indonesian names of the powers of a thousand, from
// the biggest.
var terbilangScales = []struct {
	value uint64
	name  string
}{
	{1_000_000_000_000, "triliun"},
	{1_000_000_000, "miliar"},
	{1_000_000, "juta"},
	{1_000, "ribu"},
}

// terbilang writes n in Indonesian words, the way amounts are written on
// letters, e.g. 12 is "dua belas", 101 "seratus satu" and 1000 "seribu".
// Quantities on purchase orders are at least 1, but 0 is "nol".
func terbilang(n uint64) string {
	if n == 0 {
		return terbilangDigits[0]
	}

	var words []string
	for _, scale := range terbilangScales {
		if n < scale.value {
			continue
		}

		count := n / scale.value
		n %= scale.value

		if count == 1 && scale.name == "ribu" {
			words = append(words, "seribu")
		} else {
			words = append(words, terbilangBelowThousand(count), scale.name)
		}
	}

	if n > 0 {
		words = append(words, terbilangBelowThousand(n))
	}

	return strings.Join(words, " ")
}

// terbilangBelowThousand writes 1 to 999 in words.
func terbilangBelowThousand(n uint64) string {
	var words []string

	if hundreds := n / 100; hundreds > 0 {
		if hundreds == 1 {
			words = append(words, "seratus")
		} else {
			words = append(words, terbilangDigits[hundreds], "ratus")
		}
		n %= 100
	}

	switch {
	case n == 0:
	case n < 12:
		words = append(words, terbilangDigits[n])
	case n < 20:
		words = append(words, terbilangDigits[n-10], "belas")
	default:
		words = append(words, terbilangDigits[n/10], "puluh")
		if n%10 > 0 {
			words = append(words, terbilangDigits[n%10])
		}
	}

	return strings.Join(words, " ")
}
