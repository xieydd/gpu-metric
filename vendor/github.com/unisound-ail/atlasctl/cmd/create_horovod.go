package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/unisound-ail/atlasctl/util"
	"github.com/unisound-ail/atlasctl/cli"
	"os"
	"strings"
)

func NewHorovodJobCommand() *cobra.Command {

	var (
		createArgs createHorovodJobArgs
	)

	var command = &cobra.Command{
		Use: "horovodjob",
		Short:   "Create horovodjob as training job.",
		Aliases: []string{"hj"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			/*util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = ensureNamespace(client, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}*/

			run_hj(args, &createArgs)
		},
	}

	command.Flags().StringVar(&createArgs.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&createArgs.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")

	createArgs.addCommonFlags(command)
	createArgs.addSyncFlags(command)

	command.Flags().IntVar(&createArgs.SSHPort, "sshPort", 33,
		"ssh port.")
	return command
}

type createHorovodJobArgs struct {
	SSHPort int    `yaml:"sshPort"` // --sshPort
	Cpu     string `yaml:"cpu"`     // --cpu
	Memory  string `yaml:"memory"`  // --memory

	// for common args
	createArgs `yaml:",inline"`

	// for tensorboard
	createTensorboardArgs `yaml:",inline"`

	// for sync up source code
	createSyncCodeArgs `yaml:",inline"`
}

func (createArgs *createHorovodJobArgs) prepare(args []string) (err error) {
	createArgs.Command = strings.Join(args, " ")

	err = createArgs.check()
	if err != nil {
		return err
	}

	commonArgs := &createArgs.createArgs
	err = commonArgs.transform()
	if err != nil {
		return nil
	}

	if len(envs) > 0 {
		createArgs.Envs = transformSliceToMap(envs, "=")
	}

	createArgs.addHorovodInfoToEnv()

	return nil
}

func (createArgs createHorovodJobArgs) check() error {
	err := createArgs.createArgs.check()
	if err != nil {
		return err
	}

	if createArgs.Image == "" {
		return fmt.Errorf("--image must be set ")
	}

	return nil
}

func (createArgs *createHorovodJobArgs) addHorovodInfoToEnv() {
	createArgs.addJobInfoToEnv()
}

func createHorovodJob(namespace string,args []string, createArgs *createHorovodJobArgs) (err error) {
	err = createArgs.prepare(args)
	if err != nil {
		return err
	}

	exist, err := util.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'atlasctl delete %s'", name, name)
	}

	// the master is also considered as a worker
	createArgs.WorkerCount = createArgs.WorkerCount - 1

	return util.InstallRelease(name, namespace, createArgs, horovod_training_chart)
}

func run_hj(args []string, createArgs *createHorovodJobArgs) {
	_, namespace, e := cli.GetCliSetNameSpace()
	util.Must(e)
	
	e = createHorovodJob(namespace,args, createArgs)
	util.Must(e)
}

