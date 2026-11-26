package purchaseorder

import "strings"

// drugClass is a regulatory class of drugs that has its own purchase order
// template.
type drugClass string

const (
	drugClassPrekursor drugClass = "prekursor"
	drugClassOOT       drugClass = "oot"
	// drugClassOther is any other drug class, e.g. obat keras.
	drugClassOther drugClass = "other"
)

// golonganClasses maps the vmedis golongan that are drug classes, lower
// case, to their class. Vmedis golongan also hold shelves (e.g. "DEPAN",
// "KULKAS"), which are left out, so they never count as a class.
var golonganClasses = map[string]drugClass{
	"prekursor":            drugClassPrekursor,
	"prekursor farmasi":    drugClassPrekursor,
	"obat-obatan tertentu": drugClassOOT,
	"obat-obat tertentu":   drugClassOOT,
	"obat obat tertentu":   drugClassOOT,
	"oot":                  drugClassOOT,

	"obat bebas":             drugClassOther,
	"obat bebas terbatas":    drugClassOther,
	"obat keras":             drugClassOther,
	"obat narkotika":         drugClassOther,
	"narkotika":              drugClassOther,
	"obat psikotropika":      drugClassOther,
	"psikotropika":           drugClassOther,
	"jamu":                   drugClassOther,
	"obat herbal terstandar": drugClassOther,
	"fitofarmaka":            drugClassOther,
	"alkes":                  drugClassOther,
	"alat kesehatan":         drugClassOther,
	"kosmetik":               drugClassOther,
	"suplemen":               drugClassOther,
}

// golonganClass returns the drug class of a vmedis golongan, and false when
// the golongan is not a drug class (empty, or a shelf).
func golonganClass(golongan string) (drugClass, bool) {
	class, ok := golonganClasses[strings.ToLower(strings.Join(strings.Fields(golongan), " "))]
	return class, ok
}

// golonganMatches reports whether a drug with the vmedis golongan belongs on
// the template: nil when it cannot tell, because the template is for any drug
// or the golongan is not a drug class.
func golonganMatches(template Template, golongan string) *bool {
	if template.DrugClass == "" {
		return nil
	}

	class, ok := golonganClass(golongan)
	if !ok {
		return nil
	}

	matches := class == template.DrugClass
	return &matches
}
