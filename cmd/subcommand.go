package cmd

import "github.com/spf13/cobra"

type commandWithInit struct {
	command *cobra.Command
	init    func(cmd *cobra.Command)
}

func initSubcommands(parentCommand *cobra.Command, commands []commandWithInit) {
	for _, cmd := range commands {
		parentCommand.AddCommand(cmd.command)

		if cmd.init != nil {
			cmd.init(cmd.command)
		}
	}

	initAppCommand(parentCommand)
}
