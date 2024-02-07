package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
			Short: "Run procurements one-time dumper",
			Run: func(cmd *cobra.Command, args []string) {
				db, err := getDatabase()
				if err != nil {
					log.Fatalf("Error opening database: %s\n", err)
				}

				procurement.DumpProcurementsBetweenDatesFromVmedisToDB(
					cmd.Context(),
					viper.GetTime("start_date"),
					viper.GetTime("end_date"),
					db,
					getVmedisClient(),
					getDrugProducer(),
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
}

func init() {
	for _, cmd := range procurementCommands {
		procurementsCmd.AddCommand(cmd.command)
		cmd.init(cmd.command)
	}

	initAppCommand(procurementsCmd)
}
