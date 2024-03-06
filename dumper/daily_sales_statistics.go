package dumper

import (
	"context"
	"log"

	"github.com/turfaa/vmedis-proxy-api/sale"
)

// DumpDailySalesStatistics dumps the daily sales statistics.
func DumpDailySalesStatistics(ctx context.Context, saleService *sale.Service) {
	if err := saleService.DumpTodaySalesStatisticsFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpTodaySalesStatisticsFromVmedisToDB: %s", err)
	}
}
