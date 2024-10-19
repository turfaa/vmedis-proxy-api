package shift

import (
	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func vmedisShiftToDBShift(shift vmedis.Shift) models.Shift {
	return models.Shift{
		VmedisID:            shift.ID,
		Code:                shift.Code,
		StartedAt:           shift.StartedAt.Time,
		EndedAt:             shift.EndedAt.Time,
		Cashier:             shift.Cashier,
		InitialCash:         shift.InitialCash,
		ExpectedFinalCash:   shift.ExpectedFinalCash,
		ActualFinalCash:     shift.ActualFinalCash,
		FinalCashDifference: shift.FinalCashDifference,
		Supervisor:          shift.Supervisor,
		Notes:               shift.Notes,
	}
}
