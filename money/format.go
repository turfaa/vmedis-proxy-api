package money

import "github.com/leekchan/accounting"

var (
	rupiah = accounting.Accounting{Symbol: "Rp", Format: "%s %v", FormatZero: "%s 0", Thousand: ".", Decimal: ","}
)

func FormatRupiah(amount float64) string {
	return rupiah.FormatMoney(amount)
}
