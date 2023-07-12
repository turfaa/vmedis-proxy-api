package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "vmedis-proxy-api",
	Short: "Run the vmedis proxy api server",

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		proxy.Run(
			client.New(viper.GetString("base_url"), viper.GetString("session_id")),
			viper.GetDuration("refresh_interval"),
		)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().String("base-url", "http://localhost:8080", "base url of the vmedis proxy server")
	rootCmd.PersistentFlags().String("session-id", "", "session id of the vmedis proxy server")
	rootCmd.PersistentFlags().Duration("refresh-interval", time.Minute, "refresh interval of the session id in milliseconds")

	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("session_id", rootCmd.PersistentFlags().Lookup("session-id"))
	viper.BindPFlag("refresh_interval", rootCmd.PersistentFlags().Lookup("refresh-interval"))

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
