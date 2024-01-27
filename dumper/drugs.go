package dumper

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// DumpDrugs dumps the drugs.
func DumpDrugs(ctx context.Context, db *gorm.DB, vmedisClient *vmedis.Client) {
	drugService := drug.NewService(db, vmedisClient)

	if err := drugService.DumpDrugsFromVmedisToDB(ctx); err != nil {
		log.Println("Error dumping drugs:", err)
	}
}
