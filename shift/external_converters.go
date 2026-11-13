package shift

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func vmedisShiftToDBShift(shift vmedisv2.Shift) models.Shift {
	return models.Shift{
		VmedisID:            shift.ID,
		Code:                shift.Code,
		StartedAt:           shift.StartedAt,
		EndedAt:             shift.EndedAt,
		Cashier:             shift.Cashier,
		InitialCash:         shift.InitialCash,
		ExpectedFinalCash:   shift.ExpectedFinalCash,
		ActualFinalCash:     shift.ActualFinalCash,
		FinalCashDifference: shift.FinalCashDifference,
		Supervisor:          shift.Supervisor,
		Notes:               shift.Notes,
	}
}
