package cmd

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/drug"
)

var runConsumersCmd = &cobra.Command{
	Use:   "run-consumers",
	Short: "Run consumers",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := getDatabase()
		if err != nil {
			log.Fatalf("Error opening database: %s\n", err)
		}

		drugConsumer := drug.NewConsumer(
			viper.GetStringSlice("kafka_brokers"),
			db,
			getRedisClient(),
			getVmedisClient(),
			getKafkaWriter(),
		)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			drugConsumer.StartConsuming()
			wg.Done()
		}()

		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

		<-done

		log.Println("Shutting down consumers")
		drugConsumer.Close()

		wg.Wait()
		log.Println("Consumers shut down successfully")
	},
}

func init() {
	initAppCommand(runConsumersCmd)
}
