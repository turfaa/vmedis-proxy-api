package supplier

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// SuppliersResponse is the response of the suppliers API.
type SuppliersResponse struct {
	Suppliers []Supplier `json:"suppliers"`
}

// Supplier is a supplier the pharmacy procures drugs from.
type Supplier struct {
	VmedisID          int     `json:"vmedisId"`
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	Phone             string  `json:"phone"`
	MobilePhone       string  `json:"mobilePhone"`
	BankAccountNumber string  `json:"bankAccountNumber"`
	TaxNumber         string  `json:"taxNumber"`
	Active            bool    `json:"active"`
	LeadTimeDays      int     `json:"leadTimeDays"`
	Rating            float64 `json:"rating"`
	// UpdatedAt is when the supplier was last synced from vmedis.
	UpdatedAt time.Time `json:"updatedAt"`
}

// FromDBSupplier creates Supplier from its DB model.
func FromDBSupplier(s models.Supplier) Supplier {
	return Supplier{
		VmedisID:          s.VmedisID,
		Code:              s.Code,
		Name:              s.Name,
		Address:           s.Address,
		City:              s.City,
		Phone:             s.Phone,
		MobilePhone:       s.MobilePhone,
		BankAccountNumber: s.BankAccountNumber,
		TaxNumber:         s.TaxNumber,
		Active:            s.Active,
		LeadTimeDays:      s.LeadTimeDays,
		Rating:            s.Rating,
		UpdatedAt:         s.UpdatedAt,
	}
}

// DumpStatus represents the state of the supplier dump.
type DumpStatus string

const (
	// DumpStatusIdle means no supplier dump is in progress.
	DumpStatusIdle DumpStatus = "IDLE"
	// DumpStatusDumping means suppliers are currently being dumped from Vmedis.
	DumpStatusDumping DumpStatus = "DUMPING"
)

// DumpStatusResponse is the response of the supplier dump status API.
type DumpStatusResponse struct {
	Status DumpStatus `json:"status"`
}
