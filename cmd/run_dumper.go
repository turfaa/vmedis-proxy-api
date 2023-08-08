package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

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
			client.New(
				viper.GetString("base_url"),
				viper.GetStringSlice("session_ids"),
				viper.GetInt("concurrency"),
				rate.NewLimiter(rate.Limit(viper.GetFloat64("rate_limit")), 1),
			),
			db,
		)
	},
}

func init() {
	initAppCommand(runDumperCmd)
}
