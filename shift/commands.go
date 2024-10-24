package shift

import (
	"context"
	"log"
	"time"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"gorm.io/gorm"
)

func DumpShiftsFromVmedisToDB(
	ctx context.Context,
	from time.Time,
	to time.Time,
	db *gorm.DB,
	vmedisClient *vmedis.Client,
) {
	service := NewService(db, vmedisClient)

	if _, err := service.DumpShiftsFromVmedisToDB(ctx, from, to); err != nil {
		log.Fatalf("DumpShiftsBetweenTimesFromVmedisToDB: %s", err)
	}
}
