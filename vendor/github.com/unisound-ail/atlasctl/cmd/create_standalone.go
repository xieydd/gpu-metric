package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unisound-ail/atlasctl/create"
	"os"
)


func NewStandaloneJobCommand() *cobra.Command {

	standCmd := cobra.Command{
		Use:     "standalonejob",
		Short:   "Create StandaloneJob as training job.",
		Aliases: []string{"sj"},

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			create.Run()
		},

	}
	create.AddCommonFlags(&standCmd)

	return &standCmd
}
