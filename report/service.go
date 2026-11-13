package report

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"
	"unicode/utf8"

	"github.com/jordan-wright/email"
	"github.com/xuri/excelize/v2"

	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/sale"
)

type Service struct {
	aggregatedProcurementsGetter AggregatedProcurementsGetter
	aggregatedSalesGetter        AggregatedSalesGetter
	sender                       EmailSender
}

func (s *Service) SendAggregatedProcurementsAndSalesXLSX(
	ctx context.Context,
	fromTime time.Time,
	toTime time.Time,
	from string,
	to []string,
	cc []string,
	subject string,
	body []byte,
) error {
	report, err := s.GenerateAggregatedProcurementsAndSalesXLSX(ctx, fromTime, toTime)
	if err != nil {
		return fmt.Errorf("generate aggregated procurements and sales XLSX: %w", err)
	}

	mail := email.Email{
		From:    from,
		To:      to,
		Cc:      cc,
		Subject: subject,
		Text:    body,
	}

	if _, err := mail.Attach(bytes.NewReader(report), "Laporan.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"); err != nil {
		return fmt.Errorf("attach procurements XLSX: %w", err)
	}

	if err := s.sender.Send(&mail, -1); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

// GenerateAggregatedProcurementsAndSalesXLSX builds the workbook sent to
// IQVIA: a "Pembelian" sheet with the drugs procured and a "Penjualan" sheet
// with the drugs sold between fromTime and toTime, and returns its bytes.
func (s *Service) GenerateAggregatedProcurementsAndSalesXLSX(ctx context.Context, fromTime time.Time, toTime time.Time) ([]byte, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close file: %s", err)
		}
	}()

	if err := s.GenerateAggregatedProcurementsXLSXSheet(ctx, f, fromTime, toTime); err != nil {
		return nil, fmt.Errorf("generate aggregated procurements sheet: %w", err)
	}

	if err := s.GenerateAggregatedSalesXLSXSheet(ctx, f, fromTime, toTime); err != nil {
		return nil, fmt.Errorf("generate aggregated sales sheet: %w", err)
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, fmt.Errorf("delete default XLSX sheet: %w", err)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *Service) GenerateAggregatedProcurementsXLSXSheet(ctx context.Context, file *excelize.File, from time.Time, to time.Time) error {
	procurements, err := s.aggregatedProcurementsGetter.GetAggregatedProcurementsBetweenTime(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get aggregated procurements between %s and %s: %w", from, to, err)
	}

	drugs := slices2.Map(procurements, func(p procurement.AggregatedProcurement) DrugQuantity {
		return DrugQuantity{
			DrugName: p.DrugName,
			Quantity: p.Quantity,
			Unit:     p.Unit,
		}
	})

	if err := GenerateDrugQuantitySheet(file, "Pembelian", drugs); err != nil {
		return fmt.Errorf("generate drug quantity sheet: %w", err)
	}

	return nil
}

func (s *Service) GenerateAggregatedSalesXLSXSheet(ctx context.Context, file *excelize.File, from time.Time, to time.Time) error {
	sales, err := s.aggregatedSalesGetter.GetAggregatedSalesBetweenTime(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get aggregated sales between %s and %s: %w", from, to, err)
	}

	drugs := slices2.Map(sales, func(s sale.AggregatedSale) DrugQuantity {
		return DrugQuantity{
			DrugName: s.DrugName,
			Quantity: s.Quantity,
			Unit:     s.Unit,
		}
	})

	if err := GenerateDrugQuantitySheet(file, "Penjualan", drugs); err != nil {
		return fmt.Errorf("generate drug quantity sheet: %w", err)
	}

	return nil
}

// GenerateDrugQuantitySheet writes the drugs to a new sheet: a header row with
// "Nama Obat", "Kuantitas" and "Satuan", then one row per drug.
// The layout is kept plain, the header in the first row and nothing else on
// the sheet, so that it stays easy to read automatically; only the styling
// makes it easier on the eyes.
func GenerateDrugQuantitySheet(file *excelize.File, sheetName string, drugs []DrugQuantity) error {
	if _, err := file.NewSheet(sheetName); err != nil {
		return fmt.Errorf("create excel sheet: %w", err)
	}

	styles, err := newDrugQuantityStyles(file)
	if err != nil {
		return fmt.Errorf("create styles: %w", err)
	}

	var errs []error

	errs = append(errs, file.SetSheetRow(sheetName, "A1", &[]any{"Nama Obat", "Kuantitas", "Satuan"}))
	errs = append(errs, file.SetCellStyle(sheetName, "A1", "C1", styles.header))
	errs = append(errs, file.SetRowHeight(sheetName, 1, 24))

	nameWidth := float64(len("Nama Obat"))
	for i, d := range drugs {
		row := i + 2
		errs = append(errs, file.SetCellStr(sheetName, fmt.Sprintf("A%d", row), d.DrugName))
		errs = append(errs, file.SetCellFloat(sheetName, fmt.Sprintf("B%d", row), d.Quantity, 2, 64))
		errs = append(errs, file.SetCellStr(sheetName, fmt.Sprintf("C%d", row), d.Unit))
		errs = append(errs, file.SetRowHeight(sheetName, row, 18))

		band := styles.rows[i%2]
		quantityStyle := band.decimal
		if d.Quantity == math.Trunc(d.Quantity) {
			quantityStyle = band.integer
		}
		errs = append(errs, file.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), band.text))
		errs = append(errs, file.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), quantityStyle))
		errs = append(errs, file.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), band.text))

		nameWidth = max(nameWidth, float64(utf8.RuneCountInString(d.DrugName)))
	}

	errs = append(errs, file.SetColWidth(sheetName, "A", "A", min(nameWidth+4, 70)))
	errs = append(errs, file.SetColWidth(sheetName, "B", "B", 14))
	errs = append(errs, file.SetColWidth(sheetName, "C", "C", 14))

	// Keep the header visible while scrolling, and let the columns be
	// filtered and sorted.
	errs = append(errs, file.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}))
	errs = append(errs, file.AutoFilter(sheetName, fmt.Sprintf("A1:C%d", len(drugs)+1), nil))

	return errors.Join(errs...)
}

type drugQuantityStyles struct {
	header int
	// rows are the styles of the even and odd data rows, which are shaded
	// differently.
	rows [2]drugQuantityRowStyles
}

type drugQuantityRowStyles struct {
	text    int
	integer int
	decimal int
}

const (
	reportHeaderColor = "1F4E78"
	reportBandColor   = "EEF3F8"
	reportBorderColor = "BFC9D4"
	reportFont        = "Calibri"
)

func newDrugQuantityStyles(file *excelize.File) (drugQuantityStyles, error) {
	border := []excelize.Border{
		{Type: "left", Color: reportBorderColor, Style: 1},
		{Type: "right", Color: reportBorderColor, Style: 1},
		{Type: "top", Color: reportBorderColor, Style: 1},
		{Type: "bottom", Color: reportBorderColor, Style: 1},
	}

	var (
		styles drugQuantityStyles
		errs   []error
		err    error
	)

	styles.header, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Family: reportFont, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{reportHeaderColor}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    border,
	})
	errs = append(errs, err)

	font := &excelize.Font{Family: reportFont, Size: 11}
	integerFormat, decimalFormat := "#,##0", "#,##0.00"
	for i := range styles.rows {
		var fill excelize.Fill
		if i == 1 {
			fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{reportBandColor}}
		}

		styles.rows[i].text, err = file.NewStyle(&excelize.Style{
			Font:      font,
			Fill:      fill,
			Border:    border,
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", Indent: 1},
		})
		errs = append(errs, err)

		styles.rows[i].integer, err = file.NewStyle(&excelize.Style{
			Font:         font,
			Fill:         fill,
			Border:       border,
			Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center", Indent: 1},
			CustomNumFmt: &integerFormat,
		})
		errs = append(errs, err)

		styles.rows[i].decimal, err = file.NewStyle(&excelize.Style{
			Font:         font,
			Fill:         fill,
			Border:       border,
			Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center", Indent: 1},
			CustomNumFmt: &decimalFormat,
		})
		errs = append(errs, err)
	}

	return styles, errors.Join(errs...)
}

func NewService(
	aggregatedProcurementsGetter AggregatedProcurementsGetter,
	aggregatedSalesGetter AggregatedSalesGetter,
	sender EmailSender,
) *Service {
	return &Service{
		aggregatedProcurementsGetter: aggregatedProcurementsGetter,
		aggregatedSalesGetter:        aggregatedSalesGetter,
		sender:                       sender,
	}
}
