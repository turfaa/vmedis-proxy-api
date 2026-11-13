package report

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"
)

func SendIQVIALastMonthReport(
	ctx context.Context,
	aggregatedProcurementsGetter AggregatedProcurementsGetter,
	aggregatedSalesGetter AggregatedSalesGetter,
	sender EmailSender,
	from string,
	to []string,
	cc []string,
) {
	service := NewService(aggregatedProcurementsGetter, aggregatedSalesGetter, sender)

	fromTime, toTime := time2.BeginningOfLastMonth(), time2.EndOfLastMonth()

	log.Printf("Sending last month report from %s to %s", fromTime.Format("2006-01-02"), toTime.Format("2006-01-02"))
	if err := service.SendAggregatedProcurementsAndSalesXLSX(
		ctx,
		fromTime,
		toTime,
		from,
		to,
		cc,
		fmt.Sprintf("Apotek Aulia Farma - Laporan %s", fromTime.Format("2006-01")),
		[]byte(`
Halo tim IQVIA,

Berikut adalah laporan penjualan dan pembelian per bulan dari bulan lalu yang telah kami kumpulkan.

Terima kasih.
`),
	); err != nil {
		log.Fatalf("SendAggregatedProcurementsAndSalesXLSX: %s", err)
	}

	log.Println("Last month report sent")
}

// GenerateIQVIAReport writes the report usually sent to IQVIA, covering the
// whole month that starts at month, to outputPath without sending any email.
func GenerateIQVIAReport(
	ctx context.Context,
	aggregatedProcurementsGetter AggregatedProcurementsGetter,
	aggregatedSalesGetter AggregatedSalesGetter,
	month time.Time,
	outputPath string,
) {
	service := NewService(aggregatedProcurementsGetter, aggregatedSalesGetter, nil)

	fromTime := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	toTime := fromTime.AddDate(0, 1, 0).Add(-time.Nanosecond)

	log.Printf("Generating IQVIA report from %s to %s", fromTime.Format(time.DateOnly), toTime.Format(time.DateOnly))
	report, err := service.GenerateAggregatedProcurementsAndSalesXLSX(ctx, fromTime, toTime)
	if err != nil {
		log.Fatalf("GenerateAggregatedProcurementsAndSalesXLSX: %s", err)
	}

	if err := os.WriteFile(outputPath, report, 0o644); err != nil {
		log.Fatalf("write report to %s: %s", outputPath, err)
	}

	log.Printf("IQVIA report written to %s", outputPath)
}
