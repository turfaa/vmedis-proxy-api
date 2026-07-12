package sale

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

func DumpTodaySalesFromVmedisToDB(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedisv1.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugsGetter, drugProducer)

	if err := service.DumpTodaySalesFromVmedisToDB(ctx); err != nil {
		log.Fatalf("Failed to dump today's sales from Vmedis to DB: %s", err)
	}
}

func DumpSalesByDateFromVmedisToDB(
	ctx context.Context,
	date time.Time,
	db *gorm.DB,
	vmedisClient *vmedisv1.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugsGetter, drugProducer)

	if err := service.DumpSalesByDateFromVmedisToDB(ctx, date); err != nil {
		log.Fatalf("Failed to dump sales at %s from Vmedis to DB: %s", date.Format(time.DateOnly), err)
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
