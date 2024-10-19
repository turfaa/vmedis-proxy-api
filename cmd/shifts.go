package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turfaa/vmedis-proxy-api/shift"
)

var shiftsCmd = &cobra.Command{
	Use:   "shifts",
	Short: "Shifts commands",
}

var shiftsCommands = []commandWithInit{
	{
		init: func(cmd *cobra.Command) {
			cmd.Flags().String("from", time.Now().AddDate(0, 0, -3).Format(time.DateTime), "From date")
			cmd.Flags().String("to", time.Now().Format(time.DateTime), "To date")

			viper.BindPFlag("from", cmd.Flags().Lookup("from"))
			viper.BindPFlag("to", cmd.Flags().Lookup("to"))
		},
		command: &cobra.Command{
			Use:   "dump",
			Short: "Run one-time shifts dumper",
			Run: func(cmd *cobra.Command, args []string) {
				fromUTC := viper.GetTime("from")
				from := time.Date(fromUTC.Year(), fromUTC.Month(), fromUTC.Day(), fromUTC.Hour(), fromUTC.Minute(), fromUTC.Second(), fromUTC.Nanosecond(), time.Local)

				toUTC := viper.GetTime("to")
				to := time.Date(toUTC.Year(), toUTC.Month(), toUTC.Day(), toUTC.Hour(), toUTC.Minute(), toUTC.Second(), toUTC.Nanosecond(), time.Local)

				shift.DumpShiftsFromVmedisToDB(cmd.Context(), from, to, getDatabase(), getVmedisClient())
			},
		},
	},
}

func init() {
	initSubcommands(shiftsCmd, shiftsCommands)
}
