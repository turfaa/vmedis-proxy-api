package procurement

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"
)

func DumpProcurementsBetweenDatesFromVmedisToDB(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	db *gorm.DB,
	redisClient *redis.Client,
	vmedisClient *vmedisv1.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) {
	service := NewService(db, redisClient, vmedisClient, drugProducer, drugUnitsGetter)

	if err := service.DumpProcurementsBetweenDatesFromVmedisToDB(ctx, startDate, endDate); err != nil {
		log.Fatalf("DumpProcurementsBetweenDatesFromVmedisToDB: %s", err)
	}
}

func DumpProcurementRecommendations(
	ctx context.Context,
	db *gorm.DB,
	redisClient *redis.Client,
	vmedisClient *vmedisv1.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) {
	service := NewService(db, redisClient, vmedisClient, drugProducer, drugUnitsGetter)

	if err := service.DumpRecommendationsFromVmedisToRedis(ctx); err != nil {
		log.Fatalf("DumpProcurementsBetweenDatesFromVmedisToDB: %s", err)
	}
}
