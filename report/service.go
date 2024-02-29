package report

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

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
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close file: %s", err)
		}
	}()

	if err := s.GenerateAggregatedProcurementsXLSXSheet(ctx, f, fromTime, toTime); err != nil {
		return fmt.Errorf("generate aggregated procurements sheet: %w", err)
	}

	if err := s.GenerateAggregatedSalesXLSXSheet(ctx, f, fromTime, toTime); err != nil {
		return fmt.Errorf("generate aggregated sales sheet: %w", err)
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		return fmt.Errorf("delete default XLSX sheet: %w", err)
	}

	bytesBuffer, err := f.WriteToBuffer()
	if err != nil {
		return fmt.Errorf("write to buffer: %w", err)
	}

	mail := email.Email{
		From:    from,
		To:      to,
		Cc:      cc,
		Subject: subject,
		Text:    body,
	}

	if _, err := mail.Attach(bytesBuffer, "Laporan.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"); err != nil {
		return fmt.Errorf("attach procurements XLSX: %w", err)
	}

	if err := s.sender.Send(&mail, -1); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
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

func GenerateDrugQuantitySheet(file *excelize.File, sheetName string, drugs []DrugQuantity) error {
	if _, err := file.NewSheet(sheetName); err != nil {
		return fmt.Errorf("create excel sheet: %w", err)
	}

	var errs []error

	errs = append(errs, file.SetCellStr(sheetName, "A1", "Nama Obat"))
	errs = append(errs, file.SetCellStr(sheetName, "B1", "Kuantitas"))
	errs = append(errs, file.SetCellStr(sheetName, "C1", "Satuan"))

	for i, d := range drugs {
		errs = append(errs, file.SetCellStr(sheetName, fmt.Sprintf("A%d", i+2), d.DrugName))
		errs = append(errs, file.SetCellFloat(sheetName, fmt.Sprintf("B%d", i+2), d.Quantity, 2, 64))
		errs = append(errs, file.SetCellStr(sheetName, fmt.Sprintf("C%d", i+2), d.Unit))
	}

	return errors.Join(errs...)
}

func (s *Service) SendAggregatedProcurementsAndSalesCSV(
	ctx context.Context,
	fromTime time.Time,
	toTime time.Time,
	from string,
	to []string,
	cc []string,
	subject string,
	body []byte,
) error {
	procurementsCSV, err := s.GenerateAggregatedProcurementsCSV(ctx, fromTime, toTime)
	if err != nil {
		return fmt.Errorf("generate aggregated procurements CSV: %w", err)
	}

	x, err := io.ReadAll(procurementsCSV)
	if err != nil {
		return fmt.Errorf("read aggregated procurements CSV: %w", err)
	}

	buf := bytes.NewReader(x)

	mail := email.Email{
		From:    from,
		To:      to,
		Cc:      cc,
		Subject: subject,
		Text:    body,
	}

	if _, err := mail.Attach(buf, "pembelian.csv", "text/csv"); err != nil {
		return fmt.Errorf("attach procurements CSV: %w", err)
	}

	if err := s.sender.Send(&mail, -1); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *Service) GenerateAggregatedProcurementsCSV(ctx context.Context, from time.Time, to time.Time) (io.Reader, error) {
	procurements, err := s.aggregatedProcurementsGetter.GetAggregatedProcurementsBetweenTime(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get aggregated procurements between %s and %s: %w", from, to, err)
	}

	var bytesBuffer bytes.Buffer
	writer := csv.NewWriter(&bytesBuffer)
	writer.Comma = ';'

	var errs []error
	errs = append(errs, writer.Write([]string{"Nama Obat", "Kuantitas", "Satuan"}))

	for _, p := range procurements {
		errs = append(errs, writer.Write([]string{
			p.DrugName,
			fmt.Sprintf("%.2f", p.Quantity),
			p.Unit,
		}))
	}

	writer.Flush()
	errs = append(errs, writer.Error())

	return &bytesBuffer, errors.Join(errs...)
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
