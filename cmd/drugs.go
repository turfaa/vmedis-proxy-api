package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/drug"
)

var drugsCmd = &cobra.Command{
	Use:   "drugs",
	Short: "Drugs commands",
}

var drugsCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "dump",
			Short: "Dump all drugs",
			Run: func(cmd *cobra.Command, args []string) {
				drug.DumpDrugsFromVmedisToDB(
					cmd.Context(),
					getDatabase(),
					getVmedisClient(),
					getKafkaWriter(),
				)
			},
		},
	},

	{
		command: &cobra.Command{
			Use:   "run-updated-drugs-consumer",
			Short: "Run updated drugs consumer",
			Run: func(cmd *cobra.Command, args []string) {
				drug.RunUpdatedDrugsConsumer(
					cmd.Context(),
					drug.ConsumerConfig{
						DB:           getDatabase(),
						RedisClient:  getRedisClient(),
						VmedisClient: getVmedisClient(),
						KafkaWriter:  getKafkaWriter(),
						Brokers:      viper.GetStringSlice("kafka_brokers"),
						Concurrency:  viper.GetInt("consumer_concurrency"),
					})
			},
		},
		init: func(cmd *cobra.Command) {
			cmd.Flags().Int("consumer-concurrency", 10, "Consumer concurrency")

			viper.BindPFlag("consumer_concurrency", cmd.Flags().Lookup("consumer-concurrency"))
		},
	},
}

func init() {
	initSubcommands(drugsCmd, drugsCommands)
}
