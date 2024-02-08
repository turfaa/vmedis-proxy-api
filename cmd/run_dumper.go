package cmd

import (
	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/dumper"
)

// runDumperCmd represents the runDumper command
var runDumperCmd = &cobra.Command{
	Use:   "run-dumper",
	Short: "Run data dumper",
	Run: func(cmd *cobra.Command, args []string) {
		dumper.Run(
			getVmedisClient(),
			getDatabase(),
			getRedisClient(),
			getKafkaWriter(),
		)
	},
}

func init() {
	initAppCommand(runDumperCmd)
}
