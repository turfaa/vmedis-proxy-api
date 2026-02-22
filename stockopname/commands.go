package stockopname

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

func DumpTodayStockOpnames(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedisv1.Client,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugProducer)

	if err := service.DumpTodayStockOpnamesFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpTodayStockOpnamesFromVmedisToDB: %s", err)
	}
}
