/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the vmedis proxy api server",

	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.SqliteDB(viper.GetString("sqlite_path"))
		if err != nil {
			log.Fatalf("Error opening database: %s\n", err)
		}

		proxy.Run(
			client.New(viper.GetString("base_url"), viper.GetString("session_id")),
			db,
			viper.GetDuration("refresh_interval"),
		)
	},
}

func init() {
	initAppCommand(serveCmd)
}
