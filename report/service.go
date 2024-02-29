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
)

type Service struct {
	aggregatedProcurementsGetter AggregatedProcurementsGetter
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

	if err := s.GenerateAggregatedProcurementsXLSXSheet(ctx, fromTime, toTime, f); err != nil {
		return fmt.Errorf("generate aggregated procurements sheet: %w", err)
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

func (s *Service) GenerateAggregatedProcurementsXLSXSheet(ctx context.Context, from time.Time, to time.Time, file *excelize.File) error {
	procurements, err := s.aggregatedProcurementsGetter.GetAggregatedProcurementsBetweenTime(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get aggregated procurements between %s and %s: %w", from, to, err)
	}

	if _, err := file.NewSheet("Pembelian"); err != nil {
		return fmt.Errorf("create procurement sheet: %w", err)
	}

	var errs []error

	errs = append(errs, file.SetCellStr("Pembelian", "A1", "Nama Obat"))
	errs = append(errs, file.SetCellStr("Pembelian", "B1", "Kuantitas"))
	errs = append(errs, file.SetCellStr("Pembelian", "C1", "Satuan"))

	for i, p := range procurements {
		errs = append(errs, file.SetCellStr("Pembelian", fmt.Sprintf("A%d", i+2), p.DrugName))
		errs = append(errs, file.SetCellFloat("Pembelian", fmt.Sprintf("B%d", i+2), p.Quantity, 2, 64))
		errs = append(errs, file.SetCellStr("Pembelian", fmt.Sprintf("C%d", i+2), p.Unit))
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
	sender EmailSender,
) *Service {
	return &Service{
		aggregatedProcurementsGetter: aggregatedProcurementsGetter,
		sender:                       sender,
	}
}
