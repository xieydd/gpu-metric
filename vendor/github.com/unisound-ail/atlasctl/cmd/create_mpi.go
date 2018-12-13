package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unisound-ail/atlasctl/util"
	"fmt"
	"os"
	"strings"
	"github.com/unisound-ail/atlasctl/cli"
)

var (
	mpijob_chart = "./charts/mpijob"
)

func NewMPIJObCommand() *cobra.Command {
	var (
		createArgs createMPIJobArgs
	)
	
	var command = &cobra.Command{
		Use:     "mpijob",
		Short:   "Create MPIjob as training job.",
		Aliases: []string{"mpi", "mj"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			
			util.SetLogLevel(logLevel)
			/*setupKubeconfig()
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
			
			run(args, &createArgs)
		},
	}
	
	command.Flags().StringVar(&createArgs.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&createArgs.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")
	
	// Tensorboard
	command.Flags().BoolVar(&createArgs.UseTensorboard, "tensorboard", false, "enable tensorboard")
	command.Flags().StringVar(&createArgs.TensorboardImage, "tensorboardImage", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.5.0-devel", "the docker image for tensorboard")
	command.Flags().StringVar(&createArgs.TrainingLogdir, "logdir", "/training_logs", "the training logs dir, default is /training_logs")
	
	createArgs.addCommonFlags(command)
	createArgs.addSyncFlags(command)
	
	return command
}

type createMPIJobArgs struct {
	Cpu    string   `yaml:"cpu"` 	// --cpu
	Memory string   `yaml:"memory"` // --memory
	
	// for common args
	createArgs `yaml:",inline"`
	
	// for tensorboard
	createTensorboardArgs `yaml:",inline"`
	
	// for sync up source code
	createSyncCodeArgs `yaml:",inline"`
}

func (createArgs *createMPIJobArgs) prepare(args []string) (err error) {
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
	
	err = createArgs.HandleSyncCode()
	if err != nil {
		return err
	}
	
	// enable Tensorboard
	if createArgs.UseTensorboard {
		createArgs.HostLogPath = fmt.Sprintf("/atlasctl_logs/training%s", util.RandomInt32())
	}
	
	if len(envs) > 0 {
		createArgs.Envs = transformSliceToMap(envs, "=")
	}
	
	createArgs.addMPIInfoToEnv()
	
	return nil
}

func (createArgs createMPIJobArgs) check() error {
	err := createArgs.createArgs.check()
	if err != nil {
		return err
	}
	
	if createArgs.Image == "" {
		return fmt.Errorf("--image must be set ")
	}
	
	return nil
}

func (createArgs *createMPIJobArgs) addMPIInfoToEnv() {
	createArgs.addJobInfoToEnv()
}

// Submit MPIJob
func createMPIJob(namespace string,args []string, createArgs *createMPIJobArgs) (err error) {
	
	err = createArgs.prepare(args)
	if err != nil {
		return err
	}
	
	exist, err :=util.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'atlactl delete %s'", name, name)
	}
	
	return util.InstallRelease(name, namespace, createArgs, mpijob_chart)
}

func run(args []string, createArgs *createMPIJobArgs) {
	_, namespace, e := cli.GetCliSetNameSpace()
	util.Must(e)
	e = createMPIJob(namespace,args, createArgs)
	util.Must(e)
}
