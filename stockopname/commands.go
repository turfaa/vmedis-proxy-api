package stockopname

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func DumpTodayStockOpnames(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	drugProducer UpdatedDrugProducer,
) {
	service := NewService(db, vmedisClient, drugProducer)

	if err := service.DumpTodayStockOpnamesFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpTodayStockOpnamesFromVmedisToDB: %s", err)
	}
}
