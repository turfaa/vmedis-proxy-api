package supplier

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func vmedisSupplierToDBSupplier(s vmedisv2.Supplier) models.Supplier {
	return models.Supplier{
		VmedisID:          s.ID,
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
	}
}
