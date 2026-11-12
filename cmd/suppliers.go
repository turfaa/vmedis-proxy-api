package cmd

import (
	"github.com/spf13/cobra"

	"github.com/turfaa/vmedis-proxy-api/supplier"
)

var suppliersCmd = &cobra.Command{
	Use:   "suppliers",
	Short: "Suppliers commands",
}

var suppliersCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "dump",
			Short: "Dump all suppliers from Vmedis to the DB, soft-deleting the ones no longer in Vmedis",
			Run: func(cmd *cobra.Command, args []string) {
				supplier.DumpSuppliersFromVmedisToDB(cmd.Context(), getDatabase(), getVmedisV2Client())
			},
		},
	},
}

func init() {
	initSubcommands(suppliersCmd, suppliersCommands)
}
