package cmd

import (
"os"

"github.com/unisound-ail/atlasctl/util"
"github.com/spf13/cobra"
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "delete a training job",
		Short: "Delete a training job and its associated pods",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			
			for _, jobName := range args {
				deleteTrainingJob(jobName)
			}
		},
	}
	
	return command
}

func deleteTrainingJob(jobName string) error {
	return util.DeleteRelease(jobName)
}
