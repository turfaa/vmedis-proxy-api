package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/procurement"
)

var procurementsCmd = &cobra.Command{
	Use:   "procurements",
	Short: "Procurements commands",
}

var procurementsCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "dump",
			Short: "Run one-time procurements dumper",
			Run: func(cmd *cobra.Command, args []string) {
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

				procurement.DumpProcurementsBetweenDatesFromVmedisToDB(
					cmd.Context(),
					startTime,
					endTime,
					getDatabase(),
					getRedisClient(),
					getVmedisClient(),
					getDrugProducer(),
					drug.NewDatabase(getDatabase()),
				)
			},
		},
		init: func(cmd *cobra.Command) {
			cmd.Flags().Int("days", 14, "Number of days to dump if start-date is not set")
			cmd.Flags().String("start-date", "", "Start date")
			cmd.Flags().String("end-date", time.Now().Format(time.DateOnly), "End date")

			viper.BindPFlag("days", cmd.Flags().Lookup("days"))
			viper.BindPFlag("start_date", cmd.Flags().Lookup("start-date"))
			viper.BindPFlag("end_date", cmd.Flags().Lookup("end-date"))
		},
	},
	{
		command: &cobra.Command{
			Use:   "dump-recommendations",
			Short: "Run one-time procurement recommendations dumper",
			Run: func(cmd *cobra.Command, args []string) {
				procurement.DumpProcurementRecommendations(
					cmd.Context(),
					getDatabase(),
					getRedisClient(),
					getVmedisClient(),
					getDrugProducer(),
					drug.NewDatabase(getDatabase()),
				)
			},
		},
	},
}

func init() {
	initSubcommands(procurementsCmd, procurementsCommands)
}
