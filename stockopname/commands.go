package stockopname

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func DumpStockOpnamesBetweenDates(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	db *gorm.DB,
	vmedisClient *vmedisv2.Client,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugProducer)

	if err := service.DumpStockOpnamesBetweenDatesFromVmedisToDB(ctx, startDate, endDate); err != nil {
		log.Fatalf("DumpStockOpnamesBetweenDatesFromVmedisToDB: %s", err)
	}
}
