package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/dumper"
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
			getVmedisClient(),
			db,
			getRedisClient(),
			getKafkaWriter(),
		)
	},
}

func init() {
	initAppCommand(runDumperCmd)
}
