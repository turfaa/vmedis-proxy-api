package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/report"
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Reports commands",
}

var reportsCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use: "send-to-iqvia",
			Run: func(cmd *cobra.Command, args []string) {
				report.SendIQVIALastMonthReport(
					cmd.Context(),
					getProcurementService(),
					getSaleService(),
					getEmailer(),
					viper.GetString("email.from"),
					viper.GetStringSlice("email.iqvia.to"),
					viper.GetStringSlice("email.iqvia.cc"),
				)
			},
		},
		init: func(cmd *cobra.Command) {
			cmd.Flags().StringSlice("email-to", nil, "email to")
			cmd.Flags().StringSlice("email-cc", nil, "email cc")

			viper.BindPFlag("email.iqvia.to", cmd.Flags().Lookup("email-to"))
			viper.BindPFlag("email.iqvia.cc", cmd.Flags().Lookup("email-cc"))
		},
	},
	{
		command: &cobra.Command{
			Use:   "generate-iqvia",
			Short: "Generate the monthly IQVIA report as an XLSX file without emailing it",
			Long: `Generate the monthly IQVIA report (the procurements and sales of a month)
as an XLSX file, the same one send-to-iqvia attaches to its email, without
sending anything. Defaults to last month.`,
			Args: cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				monthStr, _ := cmd.Flags().GetString("month")
				month, err := time.ParseInLocation("2006-01", monthStr, time.Local)
				if err != nil {
					log.Fatalf("invalid --month %q, expected YYYY-MM: %s", monthStr, err)
				}

				output, _ := cmd.Flags().GetString("output")
				if output == "" {
					output = fmt.Sprintf("Laporan-IQVIA-%s.xlsx", month.Format("2006-01"))
				}

				report.GenerateIQVIAReport(
					cmd.Context(),
					getProcurementService(),
					getSaleService(),
					month,
					output,
				)
			},
		},
		init: func(cmd *cobra.Command) {
			now := time.Now()
			lastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
			cmd.Flags().String("month", lastMonth.Format("2006-01"), "Month to report, as YYYY-MM")
			cmd.Flags().StringP("output", "o", "", "Output file (default \"Laporan-IQVIA-<month>.xlsx\")")
		},
	},
}

func init() {
	initSubcommands(reportsCmd, reportsCommands)
}
