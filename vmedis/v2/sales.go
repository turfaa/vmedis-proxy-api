package vmedisv2

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// Sale is a sale in vmedis.
type Sale struct {
	ID            int
	Date          time.Time
	Cashier       string
	InvoiceNumber string
	PatientName   string
	Doctor        string
	Salesman      string
	Payment       string
	Total         float64
	SaleUnits     []SaleUnit
}

// SaleUnit is one line of a sale.
type SaleUnit struct {
	// IDInSale is the 1-based position of the line in the sale.
	IDInSale      int
	DrugCode      string
	DrugName      string
	Batch         string
	Amount        float64
	Unit          string
	UnitPrice     float64
	PriceCategory string
	Discount      float64
	Tuslah        float64
	Embalase      float64
	Total         float64
}

const (
	// salesPageSize is the number of sales fetched per page.
	salesPageSize = 1000

	// saleUnitsConcurrency is the number of sales whose units are fetched
	// concurrently, one request per sale.
	saleUnitsConcurrency = 10
)

// GetAllSalesBetweenDates gets all sales between the given dates, inclusive,
// with their units. Only the date parts are used.
// It calls the endpoints the "Lap Penjualan Obat" v2 report page
// (/apt-lap-penjualanobat-batch-v2) gets its rows and their details from.
func (c *Client) GetAllSalesBetweenDates(ctx context.Context, startDate, endDate time.Time) ([]Sale, error) {
	salesmen, err := c.getSalesmen(ctx)
	if err != nil {
		return nil, fmt.Errorf("get salesmen: %w", err)
	}

	var rows []saleRow
	for offset := 0; ; offset += salesPageSize {
		page, err := c.getSalesPage(ctx, startDate, endDate, offset)
		if err != nil {
			return nil, fmt.Errorf("get sales between %s and %s at offset %d: %w", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly), offset, err)
		}

		rows = append(rows, page...)
		if len(page) < salesPageSize {
			break
		}
	}

	sales := make([]Sale, len(rows))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(saleUnitsConcurrency)
	for i, row := range rows {
		g.Go(func() error {
			sale, err := row.toSale(salesmen)
			if err != nil {
				return fmt.Errorf("convert sale %s: %w", row.InvoiceNumber, err)
			}

			sale.SaleUnits, err = c.getSaleUnits(gctx, int(row.ID))
			if err != nil {
				return fmt.Errorf("get units of sale %s: %w", row.InvoiceNumber, err)
			}

			sales[i] = sale
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return sales, nil
}

func (c *Client) getSalesPage(ctx context.Context, startDate, endDate time.Time, offset int) ([]saleRow, error) {
	var rows []saleRow
	err := c.callReport(ctx, "apt-lap-penjualanobat-batch/index-v3", map[string]any{
		"tanggalawal":  startDate.Format(time.DateOnly),
		"tanggalakhir": endDate.Format(time.DateOnly),
		"limit":        salesPageSize,
		"offset":       offset,
	}, &rows)
	return rows, err
}

func (c *Client) getSaleUnits(ctx context.Context, saleID int) ([]SaleUnit, error) {
	var rows []saleUnitRow
	if err := c.callReport(ctx, "apt-lap-penjualanobat-batch/indexdetail-v2", map[string]any{"pjid": saleID}, &rows); err != nil {
		return nil, err
	}

	// The lines are in the order they were added to the sale.
	slices.SortFunc(rows, func(a, b saleUnitRow) int { return cmp.Compare(a.ID, b.ID) })

	units := make([]SaleUnit, len(rows))
	for i, row := range rows {
		unit, err := row.toSaleUnit(i + 1)
		if err != nil {
			return nil, fmt.Errorf("convert unit %d: %w", int(row.ID), err)
		}
		units[i] = unit
	}

	return units, nil
}

// getSalesmen returns the names of the users who can be a sale's salesman,
// by their user id.
func (c *Client) getSalesmen(ctx context.Context) (map[string]string, error) {
	var rows []struct {
		UserID string `json:"user_id"`
		Name   string `json:"nama_lengkap"`
	}
	if err := c.callReport(ctx, "appenjualanobat-batch/datasales", nil, &rows); err != nil {
		return nil, err
	}

	salesmen := make(map[string]string, len(rows))
	for _, row := range rows {
		salesmen[row.UserID] = row.Name
	}

	return salesmen, nil
}

// saleRow is a row of apt-lap-penjualanobat-batch/index-v3.
type saleRow struct {
	ID            number  `json:"pjid"`
	Date          string  `json:"pjtanggal"`
	Cashier       string  `json:"username"`
	InvoiceNumber string  `json:"pjnofaktur"`
	PatientName   *string `json:"pasnama"`
	Doctor        *string `json:"doknama"`
	SalesmanID    string  `json:"pjsales"`
	Payment       string  `json:"pjjenis"`
	Total         number  `json:"grandtotal"`
}

// saleDateLayout is the layout of pjtanggal. Despite the "Z", it is the local
// time of the sale (it matches the "tgl" column the report displays).
const saleDateLayout = "2006-01-02T15:04:05.000Z"

func (r saleRow) toSale(salesmen map[string]string) (Sale, error) {
	date, err := time.ParseInLocation(saleDateLayout, r.Date, time.Local)
	if err != nil {
		return Sale{}, fmt.Errorf("parse date %q: %w", r.Date, err)
	}

	salesman := ""
	if r.SalesmanID != "" {
		var ok bool
		if salesman, ok = salesmen[r.SalesmanID]; !ok {
			log.Printf("Unknown salesman id %q in sale %s, using the id as the name", r.SalesmanID, r.InvoiceNumber)
			salesman = r.SalesmanID
		}
	}

	return Sale{
		ID:            int(r.ID),
		Date:          date,
		Cashier:       r.Cashier,
		InvoiceNumber: r.InvoiceNumber,
		PatientName:   strings.TrimSpace(deref(r.PatientName)),
		Doctor:        strings.TrimSpace(deref(r.Doctor)),
		Salesman:      salesman,
		Payment:       paymentLabel(r.Payment),
		Total:         float64(r.Total),
	}, nil
}

// paymentLabels are the payment labels the v1 report pages show, for the
// payment types whose label is not just the upper-cased type.
var paymentLabels = map[string]string{
	"platform": "PLATFORM ONLINE",
}

// paymentLabel returns the label of the payment type (pjjenis), e.g.
// "TUNAI" for "tunai", matching what the v1 report pages showed.
func paymentLabel(paymentType string) string {
	if label, ok := paymentLabels[paymentType]; ok {
		return label
	}
	return strings.ToUpper(paymentType)
}

// saleUnitRow is a row of apt-lap-penjualanobat-batch/indexdetail-v2.
type saleUnitRow struct {
	ID       number `json:"pjdid"`
	DrugCode string `json:"obatkode"`
	DrugName string `json:"pjdnama"`
	// Batches is a JSON array of strings, e.g. `[" 0489 (ED : 2028-06-01)"]`.
	Batches        string `json:"pjdnobatch"`
	Amount         number `json:"pjdjumlah"`
	Unit           string `json:"sonama"`
	UnitPrice      number `json:"pjdhargajual"`
	PriceID        string `json:"pjdhrgid"`
	PriceCategory  number `json:"pjdhrgcombo"`
	DiscountAmount number `json:"pjddiskontext"`
	Tuslah         number `json:"pjdtuslah"`
	Embalase       number `json:"pjdemba"`
	Total          number `json:"pjdtotal"`
}

// prescriptionHeaderPriceID matches the price id of the line heading a
// prescription's drugs, which has no price category.
var prescriptionHeaderPriceID = regexp.MustCompile(`^resep\d+$`)

func (r saleUnitRow) toSaleUnit(idInSale int) (SaleUnit, error) {
	batch, err := parseBatches(r.Batches)
	if err != nil {
		return SaleUnit{}, err
	}

	priceCategory := ""
	if !prescriptionHeaderPriceID.MatchString(r.PriceID) {
		priceCategory = fmt.Sprintf("Harga Jual %d", int(r.PriceCategory))
	}

	return SaleUnit{
		IDInSale:      idInSale,
		DrugCode:      r.DrugCode,
		DrugName:      strings.TrimSpace(r.DrugName),
		Batch:         batch,
		Amount:        float64(r.Amount),
		Unit:          r.Unit,
		UnitPrice:     float64(r.UnitPrice),
		PriceCategory: priceCategory,
		Discount:      float64(r.DiscountAmount),
		Tuslah:        float64(r.Tuslah),
		Embalase:      float64(r.Embalase),
		Total:         float64(r.Total),
	}, nil
}

func parseBatches(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var batches []string
	if err := json.Unmarshal([]byte(s), &batches); err != nil {
		return "", fmt.Errorf("parse batches %q: %w", s, err)
	}

	for i := range batches {
		batches[i] = strings.TrimSpace(batches[i])
	}

	return strings.Join(batches, ", "), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
