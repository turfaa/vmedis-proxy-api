package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turfaa/vmedis-proxy-api/database"
	"gorm.io/gorm"
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

	command.Flags().String("postgres-dsn", "", "postgres dsn, takes precedence over sqlite-path")
	command.Flags().String("sqlite-path", "data/db.sqlite", "path to the sqlite database")
	command.Flags().String("base-url", "http://localhost:8080", "base url of the vmedis proxy server")
	command.Flags().StringSlice("session-ids", nil, "session id of the vmedis proxy server")
	command.Flags().Duration("refresh-interval", time.Minute, "refresh interval of the session id in milliseconds")
	command.Flags().Int("concurrency", 50, "number of concurrent requests")
	command.Flags().Float64("rate-limit", 100, "rate limit of the requests per second")
	command.Flags().String("redis-address", "localhost:6379", "redis address")
	command.Flags().String("redis-password", "", "redis password")
	command.Flags().String("redis-db", "0", "redis db")

	viper.BindPFlag("postgres_dsn", command.Flags().Lookup("postgres-dsn"))
	viper.BindPFlag("sqlite_path", command.Flags().Lookup("sqlite-path"))
	viper.BindPFlag("base_url", command.Flags().Lookup("base-url"))
	viper.BindPFlag("session_ids", command.Flags().Lookup("session-ids"))
	viper.BindPFlag("refresh_interval", command.Flags().Lookup("refresh-interval"))
	viper.BindPFlag("concurrency", command.Flags().Lookup("concurrency"))
	viper.BindPFlag("rate_limit", command.Flags().Lookup("rate-limit"))
	viper.BindPFlag("redis_address", command.Flags().Lookup("redis-address"))
	viper.BindPFlag("redis_password", command.Flags().Lookup("redis-password"))
	viper.BindPFlag("redis_db", command.Flags().Lookup("redis-db"))
}

func getDatabase() (*gorm.DB, error) {
	if viper.GetString("postgres_dsn") != "" {
		return database.PostgresDB(viper.GetString("postgres_dsn"))
	}

	return database.SqliteDB(viper.GetString("sqlite_path"))
}
