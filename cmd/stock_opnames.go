package cmd

import (
	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/stockopname"
)

var stockOpnamesCmd = &cobra.Command{
	Use:   "stock-opnames",
	Short: "Stock opnames commands",
}

var stockOpnamesCommands = []cobra.Command{
	{
		Use:   "dump",
		Short: "Dump today's stock opnames",
		Run: func(cmd *cobra.Command, args []string) {
			stockopname.DumpTodayStockOpnames(
				cmd.Context(),
				getDatabase(),
				getVmedisClient(),
				getDrugProducer(),
			)
		},
	},
}

func init() {
	for _, cmd := range stockOpnamesCommands {
		stockOpnamesCmd.AddCommand(&cmd)
	}

	initAppCommand(stockOpnamesCmd)
}
