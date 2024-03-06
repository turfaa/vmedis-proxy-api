package dumper

import (
	"context"
	"log"

	"github.com/turfaa/vmedis-proxy-api/sale"
)

// DumpDailySales dumps the daily sales.
func DumpDailySales(ctx context.Context, saleService *sale.Service) {
	if err := saleService.DumpTodaySalesFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpTodaySalesFromVmedisToDB: %s", err)
	}
}
