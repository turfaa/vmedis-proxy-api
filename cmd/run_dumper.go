/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
)

// runDumperCmd represents the runDumper command
var runDumperCmd = &cobra.Command{
	Use:   "run-dumper",
	Short: "Run data dumper",
	Run: func(cmd *cobra.Command, args []string) {
		dumper.Run(
			client.New(viper.GetString("base_url"), viper.GetString("session_id")),
		)
	},
}

func init() {
	rootCmd.AddCommand(runDumperCmd)

	runDumperCmd.Flags().String("base-url", "http://localhost:8080", "base url of the vmedis proxy server")
	runDumperCmd.Flags().String("session-id", "", "session id of the vmedis proxy server")
	runDumperCmd.Flags().Duration("refresh-interval", time.Minute, "refresh interval of the session id in milliseconds")

	viper.BindPFlag("base_url", runDumperCmd.Flags().Lookup("base-url"))
	viper.BindPFlag("session_id", runDumperCmd.Flags().Lookup("session-id"))
	viper.BindPFlag("refresh_interval", runDumperCmd.Flags().Lookup("refresh-interval"))
}
