package dumper

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// DumpDrugs dumps the drugs.
func DumpDrugs(ctx context.Context, db *gorm.DB, vmedisClient *vmedis.Client, kafkaWriter *kafka.Writer) {
	drugService := drug.NewService(db, vmedisClient, kafkaWriter)

	if err := drugService.DumpDrugsFromVmedisToDB(ctx); err != nil {
		log.Println("Error dumping drugs:", err)
	}
}
