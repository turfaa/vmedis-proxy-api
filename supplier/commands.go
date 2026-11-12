package supplier

import (
	"context"
	"errors"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func DumpSuppliersFromVmedisToDB(ctx context.Context, db *gorm.DB, redisClient redis.UniversalClient, vmedisClient *vmedisv2.Client) {
	service := NewService(db, redisClient, vmedisClient)

	if err := service.DumpSuppliersFromVmedisToDB(ctx); err != nil {
		if errors.Is(err, ErrDumpInProgress) {
			log.Println("Suppliers are already being dumped by another process, skipping")
			return
		}

		log.Fatalf("DumpSuppliersFromVmedisToDB: %s", err)
	}
}
