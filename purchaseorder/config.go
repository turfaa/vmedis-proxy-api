package purchaseorder

import (
	"fmt"
	"strings"
)

// vmedisSupplierPagePath is the vmedis web page where suppliers are edited.
// Vmedis has no page per supplier, so the operator searches for it there.
const vmedisSupplierPagePath = "/apt-data-supplier-v2"

// Config holds the pharmacy details printed on purchase orders.
type Config struct {
	PharmacyName          string
	PharmacyLicenseNumber string
	PharmacyAddress       string
	PharmacyEmail         string
	PharmacyPhone         string
	// PharmacyCity is where letters are written from, printed before the date.
	PharmacyCity string

	PharmacistName string
	// PharmacistLicenseNumber is the pharmacist's SIPA.
	PharmacistLicenseNumber string

	// VmedisBaseURL is the base URL of the vmedis web app.
	VmedisBaseURL string
}

type configField struct {
	key   string
	value string
}

func (c Config) fields() []configField {
	return []configField{
		{"pharmacy_name", c.PharmacyName},
		{"pharmacy_license_number", c.PharmacyLicenseNumber},
		{"pharmacy_address", c.PharmacyAddress},
		{"pharmacy_email", c.PharmacyEmail},
		{"pharmacy_phone", c.PharmacyPhone},
		{"pharmacy_city", c.PharmacyCity},
		{"pharmacist_name", c.PharmacistName},
		{"pharmacist_license_number", c.PharmacistLicenseNumber},
	}
}

// Enabled reports whether any of the pharmacy details is set. Purchase
// orders are disabled until they are.
func (c Config) Enabled() bool {
	for _, f := range c.fields() {
		if strings.TrimSpace(f.value) != "" {
			return true
		}
	}

	return false
}

// Validate returns an error naming the empty fields when the feature is
// enabled but only partly configured.
func (c Config) Validate() error {
	if !c.Enabled() {
		return nil
	}

	var missing []string
	for _, f := range c.fields() {
		if strings.TrimSpace(f.value) == "" {
			missing = append(missing, f.key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("purchase orders are partly configured, missing: %s", strings.Join(missing, ", "))
	}

	return nil
}

// VmedisSupplierPageURL returns the vmedis page where suppliers are edited.
func (c Config) VmedisSupplierPageURL() string {
	return strings.TrimRight(c.VmedisBaseURL, "/") + vmedisSupplierPagePath
}

func (c Config) signer() Signer {
	return Signer{Name: c.PharmacistName, LicenseNumber: c.PharmacistLicenseNumber}
}
