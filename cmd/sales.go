package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
				dateStr := viper.GetString("date")
				if dateStr == "" {
					sale.DumpTodaySalesFromVmedisToDB(
						cmd.Context(),
						getDatabase(),
						getVmedisClient(),
						getDrugService(),
						getDrugProducer(),
					)
					return
				}

				date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
				if err != nil {
					panic(err)
				}

				sale.DumpSalesByDateFromVmedisToDB(
					cmd.Context(),
					date,
					getDatabase(),
					getVmedisClient(),
					getDrugService(),
					getDrugProducer(),
				)
			},
		},
		init: func(cmd *cobra.Command) {
			cmd.Flags().String("date", "", "Date (YYYY-MM-DD) of the sales to dump; defaults to today")

			viper.BindPFlag("date", cmd.Flags().Lookup("date"))
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
