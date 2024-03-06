package drug

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

func DumpDrugsFromVmedisToDB(
	ctx context.Context,
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	kafkaWriter *kafka.Writer,
) {
	service := NewService(db, vmedisClient, kafkaWriter)

	if err := service.DumpDrugsFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpDrugsFromVmedisToDB: %s", err)
	}
}
