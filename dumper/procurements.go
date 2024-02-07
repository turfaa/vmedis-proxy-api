package dumper

import (
	"context"
	"log"
	"time"

	"github.com/turfaa/vmedis-proxy-api/procurement"
)

func DumpProcurements(ctx context.Context, procurementService *procurement.Service) {
	start := time.Now().AddDate(0, 0, -14)
	end := time.Now()

	if err := procurementService.DumpProcurementsBetweenDatesFromVmedisToDB(ctx, start, end); err != nil {
		log.Println("Error dumping procurements:", err)
	}
}
