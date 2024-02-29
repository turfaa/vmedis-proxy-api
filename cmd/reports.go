package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/report"
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Reports commands",
}

var reportCommands = []commandWithInit{
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
}

func init() {
	for _, cmd := range reportCommands {
		reportsCmd.AddCommand(cmd.command)

		if cmd.init != nil {
			cmd.init(cmd.command)
		}
	}

	rootCmd.AddCommand(reportsCmd)
}
