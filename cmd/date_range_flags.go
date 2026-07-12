package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// registerDateRangeFlags registers the date-range flags shared by the
// dump commands (`days`, `start-date`, `end-date`) on the given command.
func registerDateRangeFlags(cmd *cobra.Command, defaultDays int) {
	cmd.Flags().Int("days", defaultDays, "Number of days to dump if start-date is not set")
	cmd.Flags().String("start-date", "", "Start date")
	cmd.Flags().String("end-date", time.Now().Format(time.DateOnly), "End date")
}

// getDateRangeFromFlags resolves the flags registered by registerDateRangeFlags
// into a [startTime, endTime] range: end-date is extended to the end of the day,
// and the start date falls back to `days` days before it when start-date is not set.
//
// It must be called from the executed command's Run. Multiple commands register
// these flags under the same viper keys, and every command's init runs on every
// invocation, so binding there would leave the keys pointing at whichever
// command registered last. Binding here points them at the executed command's flags.
func getDateRangeFromFlags(cmd *cobra.Command) (startTime time.Time, endTime time.Time) {
	viper.BindPFlag("days", cmd.Flags().Lookup("days"))
	viper.BindPFlag("start_date", cmd.Flags().Lookup("start-date"))
	viper.BindPFlag("end_date", cmd.Flags().Lookup("end-date"))

	endDate := viper.GetTime("end_date")
	endTime = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

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

	startTime = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	return startTime, endTime
}
