package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Tokens commands",
}

var tokensCommands = []cobra.Command{
	{
		Use:   "refresh",
		Short: "Refresh tokens",
		Run: func(cmd *cobra.Command, args []string) {
			refresher := getTokenRefresher()
			if err := refresher.RefreshTokens(cmd.Context()); err != nil {
				log.Fatal(err)
			}
		},
	},
}

func init() {
	for _, cmd := range tokensCommands {
		tokensCmd.AddCommand(&cmd)
	}
	
	initAppCommand(tokensCmd)
}
