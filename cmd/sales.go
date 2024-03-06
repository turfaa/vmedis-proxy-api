package cmd

import (
	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/sale"
)

var salesCmd = &cobra.Command{
	Use:   "sales",
	Short: "Sales commands",
}

var salesCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "dump",
			Short: "Run one-time sales dumper",
			Run: func(cmd *cobra.Command, args []string) {
				sale.DumpTodaySalesFromVmedisToDB(
					cmd.Context(),
					getDatabase(),
					getVmedisClient(),
					getDrugService(),
					getDrugProducer(),
				)
			},
		},
	},
	{
		command: &cobra.Command{
			Use:   "dump-statistics",
			Short: "Run one-time sales statistics dumper",
			Run: func(cmd *cobra.Command, args []string) {
				sale.DumpTodaySalesStatisticsFromVmedisToDB(
					cmd.Context(),
					getDatabase(),
					getVmedisClient(),
					getDrugService(),
					getDrugProducer(),
				)
			},
		},
	},
}

func init() {
	initSubcommands(salesCmd, salesCommands)
}
