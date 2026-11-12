package vmedisv2

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Supplier is a supplier in vmedis.
type Supplier struct {
	ID                int
	Code              string
	Name              string
	Address           string
	City              string
	Phone             string
	MobilePhone       string
	BankAccountNumber string
	// TaxNumber is the supplier's NPWP.
	TaxNumber    string
	Active       bool
	LeadTimeDays int
	Rating       float64
}

// suppliersPageSize is the number of suppliers fetched per page.
const suppliersPageSize = 1000

// GetAllSuppliers gets all the suppliers, active or not, from the endpoint
// the "Data Supplier" v2 page (/apt-data-supplier-v2) lists them from.
func (c *Client) GetAllSuppliers(ctx context.Context) ([]Supplier, error) {
	var suppliers []Supplier
	for page, offset := 1, 0; ; page, offset = page+1, offset+suppliersPageSize {
		log.Printf("Getting suppliers page %d (%d per page)", page, suppliersPageSize)

		// Without "supaktif", both active and inactive suppliers are listed.
		var rows []supplierRow
		err := c.callReport(ctx, "data-supplier/list", map[string]any{
			"limit":  suppliersPageSize,
			"offset": offset,
		}, &rows)
		if err != nil {
			return nil, fmt.Errorf("get suppliers page %d: %w", page, err)
		}

		log.Printf("Got %d suppliers from page %d", len(rows), page)

		for _, row := range rows {
			suppliers = append(suppliers, row.toSupplier())
		}

		if len(rows) < suppliersPageSize {
			break
		}
	}

	log.Printf("Got %d suppliers", len(suppliers))

	return suppliers, nil
}

// supplierRow is a row of data-supplier/list.
type supplierRow struct {
	ID                number  `json:"supid"`
	Code              string  `json:"supkode"`
	Name              string  `json:"supnama"`
	Address           *string `json:"supalamat"`
	City              *string `json:"supkota"`
	Phone             *string `json:"suptelp"`
	MobilePhone       *string `json:"suphp"`
	BankAccountNumber *string `json:"supnorek"`
	TaxNumber         *string `json:"supnpwp"`
	Active            number  `json:"supaktif"`
	LeadTimeDays      number  `json:"supleadtime"`
	Rating            number  `json:"supreting"`
}

func (r supplierRow) toSupplier() Supplier {
	trim := func(s *string) string { return strings.TrimSpace(deref(s)) }

	return Supplier{
		ID:                int(r.ID),
		Code:              strings.TrimSpace(r.Code),
		Name:              strings.TrimSpace(r.Name),
		Address:           trim(r.Address),
		City:              trim(r.City),
		Phone:             trim(r.Phone),
		MobilePhone:       trim(r.MobilePhone),
		BankAccountNumber: trim(r.BankAccountNumber),
		TaxNumber:         trim(r.TaxNumber),
		Active:            r.Active == 1,
		LeadTimeDays:      int(r.LeadTimeDays),
		Rating:            float64(r.Rating),
	}
}
