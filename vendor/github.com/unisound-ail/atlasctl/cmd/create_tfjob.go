package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/unisound-ail/atlasctl/cli"
	"github.com/unisound-ail/atlasctl/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	tfjob_chart = "./charts/tfjob"
)

func NewSubmitTFJobCommand() *cobra.Command {
	var (
		createArgs createTFJobArgs
	)

	createArgs.Mode = "tfjob"

	var command = &cobra.Command{
		Use:     "tfjob",
		Short:   "Create TFJob as training job.",
		Aliases: []string{"tf"},
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

			run_tfjob(args, &createArgs)
		},
	}

	createArgs.addCommonFlags(command)
	createArgs.addSyncFlags(command)

	// TFJob
	command.Flags().StringVar(&createArgs.WorkerImage, "workerImage", "", "the docker image for tensorflow workers")
	command.Flags().StringVar(&createArgs.PSImage, "psImage", "", "the docker image for tensorflow workers")
	command.Flags().IntVar(&createArgs.PSCount, "ps", 0, "the number of the parameter servers.")
	command.Flags().IntVar(&createArgs.PSPort, "psPort", 22223, "the port of the parameter server.")
	command.Flags().IntVar(&createArgs.WorkerPort, "workerPort", 22222, "the port of the worker.")
	command.Flags().StringVar(&createArgs.WorkerCpu, "workerCpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")
	command.Flags().StringVar(&createArgs.WorkerMemory, "workerMemory", "", "the memory resource to use for the worker, like 1Gi.")
	command.Flags().StringVar(&createArgs.PSCpu, "psCpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")
	command.Flags().StringVar(&createArgs.PSMemory, "psMemory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	// How to clean up Task
	command.Flags().StringVar(&createArgs.CleanPodPolicy, "cleanTaskPolicy", "Running", "How to clean tasks after Training is done, only support Running, None.")

	// Tensorboard
	command.Flags().BoolVar(&createArgs.UseTensorboard, "tensorboard", false, "enable tensorboard")
	command.Flags().StringVar(&createArgs.TensorboardImage, "tensorboardImage", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.5.0-devel", "the docker image for tensorboard")
	command.Flags().StringVar(&createArgs.TrainingLogdir, "logdir", "/training_logs", "the training logs dir, default is /training_logs")

	// command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}

type createTFJobArgs struct {
	Port           int    // --port, it's used set workerPort and PSPort if they are not set
	WorkerImage    string `yaml:"workerImage"`    // --workerImage
	WorkerPort     int    `yaml:"workerPort"`     // --workerPort
	PSPort         int    `yaml:"psPort"`         // --psPort
	PSCount        int    `yaml:"ps"`             // --ps
	PSImage        string `yaml:"psImage"`        // --psImage
	WorkerCpu      string `yaml:"workerCPU"`      // --workerCpu
	WorkerMemory   string `yaml:"workerMemory"`   // --workerMemory
	PSCpu          string `yaml:"psCPU"`          // --psCpu
	PSMemory       string `yaml:"psMemory"`       // --psMemory
	CleanPodPolicy string `yaml:"cleanPodPolicy"` // --cleanTaskPolicy
	// determine if it has gang scheduler
	HasGangScheduler bool `yaml:"hasGangScheduler"`
	// for common args
	createArgs `yaml:",inline"`

	// for tensorboard
	createTensorboardArgs `yaml:",inline"`

	// for sync up source code
	createSyncCodeArgs `yaml:",inline"`
}

func (createArgs *createTFJobArgs) prepare(args []string) (err error) {
	createArgs.Command = strings.Join(args, " ")

	err = createArgs.transform()
	if err != nil {
		return err
	}

	err = createArgs.check()
	if err != nil {
		return err
	}

	err = createArgs.HandleSyncCode()
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
	// pass the workers, gpu to environment variables
	// addTFJobInfoToEnv(createArgs)
	createArgs.addTFJobInfoToEnv()
	return nil
}

func (createArgs createTFJobArgs) check() error {
	err := createArgs.createArgs.check()
	if err != nil {
		return err
	}

	switch createArgs.CleanPodPolicy {
	case "None", "Running":
		log.Debugf("Supported cleanTaskPolicy: %s", createArgs.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", createArgs.CleanPodPolicy)
	}

	if createArgs.WorkerCount == 0 {
		return fmt.Errorf("--workers must be greater than 0")
	}

	if createArgs.WorkerImage == "" {
		return fmt.Errorf("--image or --workerImage must be set")
	}

	// distributed tensorflow should enable workerPort
	if createArgs.WorkerCount+createArgs.PSCount > 1 {
		if createArgs.WorkerPort <= 0 {
			return fmt.Errorf("--port or --workerPort must be set")
		}
	}

	if createArgs.PSCount > 0 {
		if createArgs.PSImage == "" {
			return fmt.Errorf("--image or --psImage must be set")
		}

		if createArgs.PSPort <= 0 {
			return fmt.Errorf("--port or --psPort must be set")
		}
	}

	return nil
}

func (createArgs *createTFJobArgs) transform() error {
	if createArgs.WorkerPort == 0 {
		createArgs.WorkerPort = createArgs.Port
	}

	if createArgs.WorkerImage == "" {
		createArgs.WorkerImage = createArgs.Image
	}

	if createArgs.PSCount > 0 {
		if createArgs.PSPort == 0 {
			createArgs.PSPort = createArgs.Port
		}

		if createArgs.PSImage == "" {
			createArgs.PSImage = createArgs.Image
		}
	}

	if createArgs.UseTensorboard {
		createArgs.HostLogPath = fmt.Sprintf("/atlasctl_logs/training%s", util.RandomInt32())
	}

	//check Gang scheduler
	createArgs.checkGangCapablitiesInCluster()

	return nil
}

func (createArgs *createTFJobArgs) addTFJobInfoToEnv() {
	createArgs.addJobInfoToEnv()
}

func (createArgs *createTFJobArgs) checkGangCapablitiesInCluster() {
	gangCapablity := false
	if clientset != nil {
		_, err := clientset.AppsV1beta1().Deployments(metav1.NamespaceSystem).Get(gangSchdName, metav1.GetOptions{})
		if err != nil {
			log.Debugf("Failed to find %s due to %v", gangSchdName, err)
		} else {
			log.Debugf("Found %s successfully, the gang scheduler is enabled in the cluster.", gangSchdName)
			gangCapablity = true
		}
	}

	createArgs.HasGangScheduler = gangCapablity
}

func submitTFJob(namespace string, args []string, createArgs *createTFJobArgs) (err error) {
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
	// createArgs.WorkerCount = createArgs.WorkerCount - 1

	return util.InstallRelease(name, namespace, createArgs, tfjob_chart)
}

func run_tfjob(args []string, createArgs *createTFJobArgs) {
	_, namespace, e := cli.GetCliSetNameSpace()
	util.Must(e)

	e = submitTFJob(namespace, args, createArgs)
	util.Must(e)
}
