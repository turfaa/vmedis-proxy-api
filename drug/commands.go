package drug

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

func RunConsumer(ctx context.Context, config ConsumerConfig) {
	consumer := NewConsumer(config)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		consumer.StartConsuming()
		wg.Done()
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	go func() {
		sig := <-done

		log.Printf("%s signal received, shutting down consumers", sig)
		consumer.Close()
	}()

	go func() {
		<-ctx.Done()

		log.Printf("Context done [%s], shutting down consumers", ctx.Err())
		consumer.Close()
	}()

	wg.Wait()
	log.Println("Consumers shut down successfully")
}
