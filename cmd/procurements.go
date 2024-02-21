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

type commandWithInit struct {
	command *cobra.Command
	init    func(cmd *cobra.Command)
}

var procurementCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "dump",
			Short: "Run one-time procurements dumper",
			Run: func(cmd *cobra.Command, args []string) {
				startDate := viper.GetTime("start_date")
				startTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)

				endDate := viper.GetTime("end_date")
				endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

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
			cmd.Flags().String("start-date", time.Now().AddDate(0, 0, -14).Format(time.DateOnly), "Start date")
			cmd.Flags().String("end-date", time.Now().Format(time.DateOnly), "End date")

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
	for _, cmd := range procurementCommands {
		procurementsCmd.AddCommand(cmd.command)

		if cmd.init != nil {
			cmd.init(cmd.command)
		}
	}

	initAppCommand(procurementsCmd)
}
