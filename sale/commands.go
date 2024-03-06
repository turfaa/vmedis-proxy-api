package sale

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func DumpTodaySalesFromVmedisToDB(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugsGetter, drugProducer)

	if err := service.DumpTodaySalesFromVmedisToDB(ctx); err != nil {
		log.Fatalf("Failed to dump today's sales from Vmedis to DB: %s", err)
	}
}

func DumpTodaySalesStatisticsFromVmedisToDB(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	drugsGetter DrugsGetter,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugsGetter, drugProducer)

	if err := service.DumpTodaySalesStatisticsFromVmedisToDB(ctx); err != nil {
		log.Fatalf("Failed to dump today's sales statistics from Vmedis to DB: %s", err)
	}
}
