package sale

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

func DumpSalesBetweenDatesFromVmedisToDB(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	db *gorm.DB,
	vmedisClient *vmedisv1.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugsGetter, drugProducer)

	if err := service.DumpSalesBetweenDatesFromVmedisToDB(ctx, startDate, endDate); err != nil {
		log.Fatalf("DumpSalesBetweenDatesFromVmedisToDB: %s", err)
	}
}

func DumpTodaySalesStatisticsFromVmedisToDB(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedisv1.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugsGetter, drugProducer)

	if err := service.DumpTodaySalesStatisticsFromVmedisToDB(ctx); err != nil {
		log.Fatalf("Failed to dump today's sales statistics from Vmedis to DB: %s", err)
	}
}
