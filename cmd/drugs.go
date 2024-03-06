package cmd

import (
	"github.com/spf13/cobra"

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
}

func init() {
	initSubcommands(drugsCmd, drugsCommands)
}
