package procurement

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func DumpProcurementsBetweenDatesFromVmedisToDB(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugProducer)

	if err := service.DumpProcurementsBetweenDatesFromVmedisToDB(ctx, startDate, endDate); err != nil {
		log.Fatalf("DumpProcurementsBetweenDatesFromVmedisToDB: %s", err)
	}
}
