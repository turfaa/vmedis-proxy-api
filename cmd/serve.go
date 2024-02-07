package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/proxy"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the vmedis proxy api server",

	Run: func(cmd *cobra.Command, args []string) {
		db, err := getDatabase()
		if err != nil {
			log.Fatalf("Error opening database: %s\n", err)
		}

		proxy.Run(
			proxy.Config{
				VmedisClient:           getVmedisClient(),
				DB:                     db,
				RedisClient:            getRedisClient(),
				KafkaWriter:            getKafkaWriter(),
				SessionRefreshInterval: viper.GetDuration("refresh_interval"),
			},
		)
	},
}

func init() {
	initAppCommand(serveCmd)
}
