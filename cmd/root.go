package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		viper.AddConfigPath("./config")
		viper.AddConfigPath(fmt.Sprintf("%s/.vmedis-proxy-api", home))
	}

	viper.SetEnvPrefix("VMEDIS")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initAppCommand(command *cobra.Command) {
	rootCmd.AddCommand(command)

	command.Flags().String("sqlite-path", "data/db.sqlite", "path to the sqlite database")
	command.Flags().String("base-url", "http://localhost:8080", "base url of the vmedis proxy server")
	command.Flags().String("session-id", "", "session id of the vmedis proxy server")
	command.Flags().Duration("refresh-interval", time.Minute, "refresh interval of the session id in milliseconds")

	viper.BindPFlag("sqlite_path", command.Flags().Lookup("sqlite-path"))
	viper.BindPFlag("base_url", command.Flags().Lookup("base-url"))
	viper.BindPFlag("session_id", command.Flags().Lookup("session-id"))
	viper.BindPFlag("refresh_interval", command.Flags().Lookup("refresh-interval"))
}
