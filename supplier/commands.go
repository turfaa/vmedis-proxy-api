package supplier

import (
	"context"
	"log"

	"gorm.io/gorm"

	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

func DumpSuppliersFromVmedisToDB(ctx context.Context, db *gorm.DB, vmedisClient *vmedisv2.Client) {
	service := NewService(db, vmedisClient)

	if err := service.DumpSuppliersFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpSuppliersFromVmedisToDB: %s", err)
	}
}
