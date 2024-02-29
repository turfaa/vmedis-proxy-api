package report

import (
	"context"
	"fmt"
	"log"

	"github.com/turfaa/vmedis-proxy-api/time2"
)

func SendIQVIALastMonthReport(
	ctx context.Context,
	aggregatedProcurementsGetter AggregatedProcurementsGetter,
	sender EmailSender,
	from string,
	to []string,
	cc []string,
) {
	service := NewService(aggregatedProcurementsGetter, sender)

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
