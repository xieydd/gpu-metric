package cmd

import (
	"github.com/spf13/cobra"
)

var (
	topLong = ` Display Resource (GPU) usage.

Available Commands:
node        Display Resource (GPU) usage of nodes
job         Display Resource (GPU) usage of pods
	`
)

func NewTopCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "top",
		Short: "Display Resource (GPU) usage.",
		Long:  topLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	// create subcommands of atlactl top
	command.AddCommand(NewTopNodeCommand())
	command.AddCommand(NewTopJobCommand())

	return command
}
