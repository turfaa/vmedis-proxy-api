package cmd

import (
	"github.com/spf13/cobra"

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
				startTime, endTime := getDateRangeFromFlags(cmd)

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
			registerDateRangeFlags(cmd, 14)
		},
	},
	{
		command: &cobra.Command{
			Use:   "reconcile",
			Short: "Soft-delete procurements that no longer exist in Vmedis, one date at a time",
			Run: func(cmd *cobra.Command, args []string) {
				startTime, endTime := getDateRangeFromFlags(cmd)

				procurement.ReconcileProcurementsBetweenDatesWithVmedis(
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
			registerDateRangeFlags(cmd, 0)
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
