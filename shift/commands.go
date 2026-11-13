package shift

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"

	"gorm.io/gorm"
)

func DumpShiftsFromVmedisToDB(
	ctx context.Context,
	from time.Time,
	to time.Time,
	db *gorm.DB,
	redisClient redis.UniversalClient,
	vmedisClient *vmedisv2.Client,
) {
	service := NewService(db, redisClient, vmedisClient)

	if err := service.DumpShiftsFromVmedisToDB(ctx, from, to); err != nil {
		log.Fatalf("DumpShiftsBetweenTimesFromVmedisToDB: %s", err)
	}
}
