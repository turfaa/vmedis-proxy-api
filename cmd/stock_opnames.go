package cmd

import (
	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/stockopname"
)

var stockOpnamesCmd = &cobra.Command{
	Use:   "stock-opnames",
	Short: "Stock opnames commands",
}

var stockOpnamesCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "dump",
			Short: "Dump stock opnames, today's by default",
			Run: func(cmd *cobra.Command, args []string) {
				startTime, endTime := getDateRangeFromFlags(cmd)

				stockopname.DumpStockOpnamesBetweenDates(
					cmd.Context(),
					startTime,
					endTime,
					getDatabase(),
					getVmedisV2Client(),
					getDrugProducer(),
				)
			},
		},
		init: func(cmd *cobra.Command) {
			registerDateRangeFlags(cmd, 0)
		},
	},
}

func init() {
	initSubcommands(stockOpnamesCmd, stockOpnamesCommands)
}
