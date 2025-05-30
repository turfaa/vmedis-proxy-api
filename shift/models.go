package shift

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/money"
	"github.com/turfaa/vmedis-proxy-api/shift/templates"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Shift struct {
	VmedisID            int       `json:"vmedis_id"`
	Code                string    `json:"code"`
	Cashier             string    `json:"cashier"`
	StartedAt           time.Time `json:"started_at"`
	EndedAt             time.Time `json:"ended_at"`
	InitialCash         float64   `json:"initial_cash"`
	ExpectedFinalCash   float64   `json:"expected_final_cash"`
	ActualFinalCash     float64   `json:"actual_final_cash"`
	FinalCashDifference float64   `json:"final_cash_difference"`
	Supervisor          string    `json:"supervisor"`
	Notes               string    `json:"notes"`
}

func (s Shift) ToTemplateData() templates.ShiftData {
	return templates.ShiftData{
		Code:                s.Code,
		Cashier:             s.Cashier,
		StartedAt:           s.StartedAt.Format("02/01/2006 15:04"),
		EndedAt:             s.EndedAt.Format("02/01/2006 15:04"),
		InitialCash:         money.FormatRupiah(s.InitialCash),
		ExpectedFinalCash:   money.FormatRupiah(s.ExpectedFinalCash),
		ActualFinalCash:     money.FormatRupiah(s.ActualFinalCash),
		FinalCashDifference: money.FormatRupiah(s.FinalCashDifference),
		Supervisor:          s.Supervisor,
		Notes:               s.Notes,
	}
}

func ShiftFromDB(dbShift models.Shift) Shift {
	return Shift{
		VmedisID:            dbShift.VmedisID,
		Code:                dbShift.Code,
		Cashier:             dbShift.Cashier,
		StartedAt:           dbShift.StartedAt,
		EndedAt:             dbShift.EndedAt,
		InitialCash:         dbShift.InitialCash,
		ExpectedFinalCash:   dbShift.ExpectedFinalCash,
		ActualFinalCash:     dbShift.ActualFinalCash,
		FinalCashDifference: dbShift.FinalCashDifference,
		Supervisor:          dbShift.Supervisor,
		Notes:               dbShift.Notes,
	}
}

func ShiftFromVmedis(shift vmedis.Shift) Shift {
	return Shift{
		VmedisID:            shift.ID,
		Code:                shift.Code,
		Cashier:             shift.Cashier,
		StartedAt:           shift.StartedAt.Time,
		EndedAt:             shift.EndedAt.Time,
		InitialCash:         shift.InitialCash,
		ExpectedFinalCash:   shift.ExpectedFinalCash,
		ActualFinalCash:     shift.ActualFinalCash,
		FinalCashDifference: shift.FinalCashDifference,
		Supervisor:          shift.Supervisor,
		Notes:               shift.Notes,
	}
}
