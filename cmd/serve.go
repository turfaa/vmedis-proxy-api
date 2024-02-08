package cmd

import (
	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/proxy"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the vmedis proxy api server",

	Run: func(cmd *cobra.Command, args []string) {
		proxy.Run(
			proxy.Config{
				VmedisClient: getVmedisClient(),
				DB:           getDatabase(),
				RedisClient:  getRedisClient(),
				KafkaWriter:  getKafkaWriter(),
			},
		)
	},
}

func init() {
	initAppCommand(serveCmd)
}
