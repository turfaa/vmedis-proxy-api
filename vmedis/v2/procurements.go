package vmedisv2

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// Procurement is a procurement (a purchase invoice) in vmedis.
type Procurement struct {
	ID int

	// Date is the date of the invoice.
	Date time.Time

	// InputTime is when the procurement was input to vmedis.
	InputTime time.Time

	InvoiceNumber          string
	Supplier               string
	Warehouse              string
	PaymentType            string
	PaymentAccount         string
	Operator               string
	CashDiscountPercentage float64
	DiscountPercentage     float64
	DiscountAmount         float64
	TaxPercentage          float64
	TaxAmount              float64
	MiscellaneousCost      float64
	Total                  float64
	ProcurementUnits       []ProcurementUnit
}

// ProcurementUnit is one line of a procurement.
type ProcurementUnit struct {
	// IDInProcurement is the 1-based position of the line in the procurement.
	IDInProcurement         int
	DrugCode                string
	DrugName                string
	Amount                  float64
	Unit                    string
	UnitBasePrice           float64
	DiscountPercentage      float64
	DiscountTwoPercentage   float64
	DiscountThreePercentage float64
	TotalUnitPrice          float64
	UnitTaxedPrice          float64
	ExpiryDate              time.Time
	BatchNumber             string
	Total                   float64
}

const (
	// procurementsPageSize is the number of procurements fetched per page.
	procurementsPageSize = 1000

	// procurementUnitsConcurrency is the number of procurements whose units
	// are fetched concurrently, one request per procurement.
	procurementUnitsConcurrency = 10
)

// GetAllProcurementsBetweenDates gets all procurements whose invoice dates
// are between the given dates, inclusive, with their units. Only the date
// parts are used.
// It calls the endpoints the "Lap Pembelian Obat" v2 report page
// (/apt-lap-pembelian-obat-v2) gets its rows and their details from.
func (c *Client) GetAllProcurementsBetweenDates(ctx context.Context, startDate, endDate time.Time) ([]Procurement, error) {
	dates := fmt.Sprintf("%s - %s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))

	var rows []procurementRow
	for page, offset := 1, 0; ; page, offset = page+1, offset+procurementsPageSize {
		log.Printf("Getting procurements %s page %d (%d per page)", dates, page, procurementsPageSize)

		pageRows, err := c.getProcurementsPage(ctx, startDate, endDate, offset)
		if err != nil {
			return nil, fmt.Errorf("get procurements %s page %d: %w", dates, page, err)
		}

		log.Printf("Got %d procurements from %s page %d", len(pageRows), dates, page)

		rows = append(rows, pageRows...)
		if len(pageRows) < procurementsPageSize {
			break
		}
	}

	log.Printf("Got %d procurements %s, getting their units", len(rows), dates)

	procurements := make([]Procurement, len(rows))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(procurementUnitsConcurrency)
	for i, row := range rows {
		g.Go(func() error {
			procurement, err := row.toProcurement()
			if err != nil {
				return fmt.Errorf("convert procurement %s: %w", row.InvoiceNumber, err)
			}

			log.Printf("Getting units of procurement %d/%d (%s)", i+1, len(rows), row.InvoiceNumber)

			procurement.ProcurementUnits, err = c.getProcurementUnits(gctx, int(row.ID))
			if err != nil {
				return fmt.Errorf("get units of procurement %s: %w", row.InvoiceNumber, err)
			}

			log.Printf("Got %d units of procurement %d/%d (%s)", len(procurement.ProcurementUnits), i+1, len(rows), row.InvoiceNumber)

			procurements[i] = procurement
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	log.Printf("Got %d procurements %s with their units", len(procurements), dates)

	return procurements, nil
}

func (c *Client) getProcurementsPage(ctx context.Context, startDate, endDate time.Time, offset int) ([]procurementRow, error) {
	var res struct {
		Data []procurementRow `json:"data"`
	}
	err := c.callReport(ctx, "laporan-transaksi-pembelian-obat-batch/index-v2", map[string]any{
		"app_jenis":    1,
		"tanggalawal":  startDate.Format(time.DateOnly),
		"tanggalakhir": endDate.Format(time.DateOnly),
		"limit":        procurementsPageSize,
		"offset":       offset,
	}, &res)
	return res.Data, err
}

func (c *Client) getProcurementUnits(ctx context.Context, procurementID int) ([]ProcurementUnit, error) {
	var res struct {
		Data []procurementUnitRow `json:"data"`
	}
	if err := c.callReport(ctx, "laporan-transaksi-pembelian-obat-batch/detail", map[string]any{"id": procurementID}, &res); err != nil {
		return nil, err
	}

	// The lines are in the order they were added to the procurement.
	slices.SortFunc(res.Data, func(a, b procurementUnitRow) int { return cmp.Compare(a.ID, b.ID) })

	units := make([]ProcurementUnit, len(res.Data))
	for i, row := range res.Data {
		unit, err := row.toProcurementUnit(i + 1)
		if err != nil {
			return nil, fmt.Errorf("convert unit %d: %w", int(row.ID), err)
		}
		units[i] = unit
	}

	return units, nil
}

// procurementRow is a row of laporan-transaksi-pembelian-obat-batch/index-v2.
type procurementRow struct {
	ID                     number  `json:"pemoid"`
	Date                   string  `json:"pemotanggal"`
	InputTime              string  `json:"pemotglinput"`
	InvoiceNumber          string  `json:"pemofaktur"`
	Supplier               string  `json:"supnama"`
	Warehouse              string  `json:"gudnama"`
	PaymentType            string  `json:"pemojenis"`
	PaymentAccount         *string `json:"aknama"`
	Operator               string  `json:"username"`
	CashDiscountPercentage number  `json:"pemodiskontunai"`
	DiscountPercentage     number  `json:"pemodiskon"`
	DiscountAmount         number  `json:"pemodiskontext"`
	TaxPercentage          number  `json:"pemopajak"`
	TaxAmount              number  `json:"pemopajaktext"`
	MiscellaneousCost      number  `json:"pemobiaya"`
	Total                  number  `json:"pemograndtotal"`
}

// procurementTimeLayout is the layout of pemotanggal and pemotglinput, which
// are local times.
const procurementTimeLayout = time.DateTime

func (r procurementRow) toProcurement() (Procurement, error) {
	dateTime, err := time.ParseInLocation(procurementTimeLayout, r.Date, time.Local)
	if err != nil {
		return Procurement{}, fmt.Errorf("parse date %q: %w", r.Date, err)
	}

	inputTime, err := time.ParseInLocation(procurementTimeLayout, r.InputTime, time.Local)
	if err != nil {
		return Procurement{}, fmt.Errorf("parse input time %q: %w", r.InputTime, err)
	}

	return Procurement{
		ID:                     int(r.ID),
		Date:                   time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, time.Local),
		InputTime:              inputTime,
		InvoiceNumber:          strings.TrimSpace(r.InvoiceNumber),
		Supplier:               strings.TrimSpace(r.Supplier),
		Warehouse:              r.Warehouse,
		PaymentType:            r.PaymentType,
		PaymentAccount:         deref(r.PaymentAccount),
		Operator:               r.Operator,
		CashDiscountPercentage: float64(r.CashDiscountPercentage),
		DiscountPercentage:     float64(r.DiscountPercentage),
		DiscountAmount:         float64(r.DiscountAmount),
		TaxPercentage:          float64(r.TaxPercentage),
		TaxAmount:              float64(r.TaxAmount),
		MiscellaneousCost:      float64(r.MiscellaneousCost),
		Total:                  float64(r.Total),
	}, nil
}

// procurementUnitRow is a row of laporan-transaksi-pembelian-obat-batch/detail.
type procurementUnitRow struct {
	ID                      number `json:"pemdoid"`
	DrugCode                string `json:"obatkode"`
	DrugName                string `json:"obatnama"`
	Amount                  number `json:"pemdojumlah"`
	Unit                    string `json:"sonama"`
	UnitBasePrice           number `json:"pemdoharga"`
	DiscountPercentage      number `json:"pemdodiskon"`
	DiscountTwoPercentage   number `json:"pemdodiskon2"`
	DiscountThreePercentage number `json:"pemdodiskon3"`
	TotalUnitPrice          number `json:"pemdohpp"`
	UnitTaxedPrice          number `json:"pemdohnappn"`
	// ExpiryDate is like "2028-01-01T00:00:00.000Z". Despite the "Z", it is a
	// local date.
	ExpiryDate  *string `json:"pemdoexpired"`
	BatchNumber string  `json:"pemdonobatch"`
	Total       number  `json:"pemdototal"`
}

func (r procurementUnitRow) toProcurementUnit(idInProcurement int) (ProcurementUnit, error) {
	var expiryDate time.Time
	if r.ExpiryDate != nil && *r.ExpiryDate != "" {
		var err error
		expiryDate, err = time.ParseInLocation(saleDateLayout, *r.ExpiryDate, time.Local)
		if err != nil {
			return ProcurementUnit{}, fmt.Errorf("parse expiry date %q: %w", *r.ExpiryDate, err)
		}
	}

	return ProcurementUnit{
		IDInProcurement:         idInProcurement,
		DrugCode:                r.DrugCode,
		DrugName:                strings.TrimSpace(r.DrugName),
		Amount:                  float64(r.Amount),
		Unit:                    r.Unit,
		UnitBasePrice:           float64(r.UnitBasePrice),
		DiscountPercentage:      float64(r.DiscountPercentage),
		DiscountTwoPercentage:   float64(r.DiscountTwoPercentage),
		DiscountThreePercentage: float64(r.DiscountThreePercentage),
		TotalUnitPrice:          float64(r.TotalUnitPrice),
		UnitTaxedPrice:          float64(r.UnitTaxedPrice),
		ExpiryDate:              expiryDate,
		BatchNumber:             strings.TrimSpace(r.BatchNumber),
		Total:                   float64(r.Total),
	}, nil
}
