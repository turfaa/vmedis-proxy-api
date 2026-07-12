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
				// The `procurements dump` command binds these keys to its own
				// flags at init time; bind at run time so the two commands
				// don't clobber each other's bindings.
				viper.BindPFlag("days", cmd.Flags().Lookup("days"))
				viper.BindPFlag("start_date", cmd.Flags().Lookup("start-date"))
				viper.BindPFlag("end_date", cmd.Flags().Lookup("end-date"))

				endDate := viper.GetTime("end_date")
				endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

				var startDate time.Time

				startDateStr := viper.GetString("start_date")
				if startDateStr == "" {
					days := viper.GetInt("days")
					startDate = endTime.AddDate(0, 0, -days)
				} else {
					var err error
					startDate, err = time.Parse(time.DateOnly, startDateStr)
					if err != nil {
						panic(err)
					}
				}

				startTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)

				sale.DumpSalesBetweenDatesFromVmedisToDB(
					cmd.Context(),
					startTime,
					endTime,
					getDatabase(),
					getVmedisClient(),
					getDrugService(),
					getDrugProducer(),
				)
			},
		},
		init: func(cmd *cobra.Command) {
			cmd.Flags().Int("days", 0, "Number of days to dump if start-date is not set")
			cmd.Flags().String("start-date", "", "Start date")
			cmd.Flags().String("end-date", time.Now().Format(time.DateOnly), "End date")
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
