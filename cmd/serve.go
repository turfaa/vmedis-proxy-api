/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the vmedis proxy api server",

	Run: func(cmd *cobra.Command, args []string) {
		proxy.Run(
			client.New(viper.GetString("base_url"), viper.GetString("session_id")),
			viper.GetDuration("refresh_interval"),
		)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("base-url", "http://localhost:8080", "base url of the vmedis proxy server")
	serveCmd.Flags().String("session-id", "", "session id of the vmedis proxy server")
	serveCmd.Flags().Duration("refresh-interval", time.Minute, "refresh interval of the session id in milliseconds")

	viper.BindPFlag("base_url", serveCmd.Flags().Lookup("base-url"))
	viper.BindPFlag("session_id", serveCmd.Flags().Lookup("session-id"))
	viper.BindPFlag("refresh_interval", serveCmd.Flags().Lookup("refresh-interval"))
}
