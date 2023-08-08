package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
)

// runDumperCmd represents the runDumper command
var runDumperCmd = &cobra.Command{
	Use:   "run-dumper",
	Short: "Run data dumper",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.SqliteDB(viper.GetString("sqlite_path"))
		if err != nil {
			log.Fatalf("Error opening database: %s\n", err)
		}

		dumper.Run(
			client.New(viper.GetString("base_url"), viper.GetStringSlice("session_ids"), viper.GetInt("concurrency")),
			db,
		)
	},
}

func init() {
	initAppCommand(runDumperCmd)
}
