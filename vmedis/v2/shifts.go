package vmedisv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// Shift is a cashier shift in vmedis.
type Shift struct {
	ID        int
	Code      string
	Cashier   string
	StartedAt time.Time

	// EndedAt is zero while the shift is still open.
	EndedAt time.Time

	InitialCash float64

	// ExpectedFinalCash is the initial cash plus the cash that came in during
	// the shift minus the cash that went out, computed the way the v2 report
	// page computes its "Saldo Akhir".
	ExpectedFinalCash float64

	// ActualFinalCash is the cash counted in the register when the shift ended.
	ActualFinalCash float64

	// FinalCashDifference is ActualFinalCash minus ExpectedFinalCash.
	FinalCashDifference float64

	// Supervisor is who the shift was handed over to.
	Supervisor string

	Notes string
}

// shiftsPageSize is the number of shifts fetched per page.
const shiftsPageSize = 500

// shiftsQueryTimeLayout is the layout of the tglAwal and tglAkhir filters,
// e.g. "10 Nov 2026 00:00".
const shiftsQueryTimeLayout = "02 Jan 2006 15:04"

// GetAllShiftsBetweenTimes gets all the shifts that started between the given
// times, inclusive, to the minute.
// It calls the endpoint the "Laporan Pergantian Shift" v2 report page
// (/apt-lap-pergantian-shift-v2) gets its rows from.
func (c *Client) GetAllShiftsBetweenTimes(ctx context.Context, startTime, endTime time.Time) ([]Shift, error) {
	times := fmt.Sprintf("%s - %s", startTime.Format(shiftsQueryTimeLayout), endTime.Format(shiftsQueryTimeLayout))

	var shifts []Shift
	for page, offset := 1, 0; ; page, offset = page+1, offset+shiftsPageSize {
		log.Printf("Getting shifts %s page %d (%d per page)", times, page, shiftsPageSize)

		rows, err := c.getShiftsPage(ctx, startTime, endTime, offset)
		if err != nil {
			return nil, fmt.Errorf("get shifts %s page %d: %w", times, page, err)
		}

		log.Printf("Got %d shifts from %s page %d", len(rows), times, page)

		for _, row := range rows {
			shift, err := row.toShift()
			if err != nil {
				return nil, fmt.Errorf("convert shift %s: %w", row.Code, err)
			}
			shifts = append(shifts, shift)
		}

		if len(rows) < shiftsPageSize {
			break
		}
	}

	log.Printf("Got %d shifts %s", len(shifts), times)

	return shifts, nil
}

func (c *Client) getShiftsPage(ctx context.Context, startTime, endTime time.Time, offset int) ([]shiftRow, error) {
	var res struct {
		Data []shiftRow `json:"data"`
	}
	err := c.callReport(ctx, "aplaporangantishift/index", map[string]any{
		// cari 4 filters by tglAwal and tglAkhir.
		"cari":     4,
		"tglAwal":  startTime.Format(shiftsQueryTimeLayout),
		"tglAkhir": endTime.Format(shiftsQueryTimeLayout),
		// Without a stable order, the pages could overlap or skip shifts.
		"sorting": "shfid ASC",
		"limit":   shiftsPageSize,
		"offset":  offset,
	}, &res)
	return res.Data, err
}

// shiftRow is a row of aplaporangantishift/index.
type shiftRow struct {
	ID        number  `json:"shfid"`
	Code      string  `json:"shfkode"`
	Cashier   string  `json:"username"`
	StartedAt string  `json:"shfbuka"`
	EndedAt   *string `json:"shftutup"`

	InitialCash     number `json:"shfawal"`
	ActualFinalCash number `json:"shfakhir"`

	// Cash in.
	Sales              number `json:"shfpenjualan"`
	ReceivablePayments number `json:"shfpembayaranpiutang"`
	ClinicPayments     number `json:"shfpembayaranklinik"`
	ClinicReceivables  number `json:"shfpiutangklinik"`
	IncomingJournals   number `json:"shfjurnalmasuk"`

	// Cash out.
	SaleReturns      number `json:"shfreturpenjualan"`
	ClinicReturns    number `json:"shfreturklinik"`
	OutgoingJournals number `json:"shfjurnalkeluar"`

	Supervisor *string `json:"diserahkan"`
	Notes      *string `json:"shfnote"`
}

// shiftTimeLayout is the layout of shfbuka and shftutup, which are local
// times.
const shiftTimeLayout = time.DateTime

func (r shiftRow) toShift() (Shift, error) {
	startedAt, err := time.ParseInLocation(shiftTimeLayout, r.StartedAt, time.Local)
	if err != nil {
		return Shift{}, fmt.Errorf("parse start time %q: %w", r.StartedAt, err)
	}

	var endedAt time.Time
	if r.EndedAt != nil && *r.EndedAt != "" {
		endedAt, err = time.ParseInLocation(shiftTimeLayout, *r.EndedAt, time.Local)
		if err != nil {
			return Shift{}, fmt.Errorf("parse end time %q: %w", *r.EndedAt, err)
		}
	}

	cashIn := r.Sales + r.ReceivablePayments + r.ClinicPayments + r.ClinicReceivables + r.IncomingJournals
	cashOut := r.SaleReturns + r.ClinicReturns + r.OutgoingJournals
	expectedFinalCash := float64(r.InitialCash + cashIn - cashOut)

	return Shift{
		ID:                  int(r.ID),
		Code:                strings.TrimSpace(r.Code),
		Cashier:             strings.TrimSpace(r.Cashier),
		StartedAt:           startedAt,
		EndedAt:             endedAt,
		InitialCash:         float64(r.InitialCash),
		ExpectedFinalCash:   expectedFinalCash,
		ActualFinalCash:     float64(r.ActualFinalCash),
		FinalCashDifference: float64(r.ActualFinalCash) - expectedFinalCash,
		Supervisor:          strings.TrimSpace(deref(r.Supervisor)),
		Notes:               strings.TrimSpace(deref(r.Notes)),
	}, nil
}
