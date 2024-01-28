package cmd

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
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
			redis.NewClient(&redis.Options{
				Addr:     viper.GetString("redis_address"),
				Password: viper.GetString("redis_password"),
				DB:       viper.GetInt("redis_db"),
			}),
			vmedis.New(
				viper.GetString("base_url"),
				viper.GetStringSlice("session_ids"),
				viper.GetInt("concurrency"),
				rate.NewLimiter(rate.Limit(viper.GetFloat64("rate_limit")), 1),
			),
			&kafka.Writer{
				Addr:         kafka.TCP(viper.GetStringSlice("kafka_brokers")...),
				Balancer:     &kafka.LeastBytes{},
				RequiredAcks: kafka.RequireOne,
				Compression:  kafka.Snappy,
			},
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
