package cmd

import (
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/turfaa/vmedis-proxy-api/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// runDumperCmd represents the runDumper command
var runDumperCmd = &cobra.Command{
	Use:   "run-dumper",
	Short: "Run data dumper",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := getDatabase()
		if err != nil {
			log.Fatalf("Error opening database: %s\n", err)
		}

		dumper.Run(
			vmedis.New(
				viper.GetString("base_url"),
				viper.GetStringSlice("session_ids"),
				viper.GetInt("concurrency"),
				rate.NewLimiter(rate.Limit(viper.GetFloat64("rate_limit")), 1),
			),
			db,
			redis.NewClient(&redis.Options{
				Addr:     viper.GetString("redis_address"),
				Password: viper.GetString("redis_password"),
				DB:       viper.GetInt("redis_db"),
			}),
			&kafka.Writer{
				Addr:         kafka.TCP(viper.GetStringSlice("kafka_brokers")...),
				Balancer:     &kafka.LeastBytes{},
				RequiredAcks: kafka.RequireOne,
				Compression:  kafka.Snappy,
			},
		)
	},
}

func init() {
	initAppCommand(runDumperCmd)
}
