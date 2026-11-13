package procurement

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func DumpProcurementsBetweenDatesFromVmedisToDB(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	db *gorm.DB,
	redisClient redis.UniversalClient,
	vmedisV2Client *vmedisv2.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) {
	service := NewService(db, redisClient, vmedisV2Client, drugProducer, drugUnitsGetter)

	if err := service.DumpProcurementsBetweenDatesFromVmedisToDB(ctx, startDate, endDate); err != nil {
		log.Fatalf("DumpProcurementsBetweenDatesFromVmedisToDB: %s", err)
	}
}

func ReconcileProcurementsBetweenDatesWithVmedis(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	db *gorm.DB,
	redisClient redis.UniversalClient,
	vmedisV2Client *vmedisv2.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) {
	service := NewService(db, redisClient, vmedisV2Client, drugProducer, drugUnitsGetter)

	if err := service.ReconcileProcurementsBetweenDatesWithVmedis(ctx, startDate, endDate); err != nil {
		log.Fatalf("ReconcileProcurementsBetweenDatesWithVmedis: %s", err)
	}
}

func DumpProcurementRecommendations(
	ctx context.Context,
	db *gorm.DB,
	redisClient redis.UniversalClient,
	vmedisV2Client *vmedisv2.Client,
	drugProducer UpdatedDrugProducer,
	drugUnitsGetter DrugUnitsGetter,
) {
	service := NewService(db, redisClient, vmedisV2Client, drugProducer, drugUnitsGetter)

	if err := service.DumpRecommendationsFromVmedisToRedis(ctx); err != nil {
		log.Fatalf("DumpProcurementsBetweenDatesFromVmedisToDB: %s", err)
	}
}
