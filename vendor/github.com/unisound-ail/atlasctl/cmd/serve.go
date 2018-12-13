package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	validate "github.com/unisound-ail/atlasctl/util"
	"github.com/spf13/cobra"
	"regexp"
	log "github.com/sirupsen/logrus"
)

var (
	modelPathSeparator = ":"
	regexp4serviceName = "^[a-z0-9A-Z_-]+$"
)

type ServeArgs struct {
	Image          string            `yaml:"image"`          // --image
	Gpus           int               `yaml:"gpus"`           // --gpus
	Cpu            string            `yaml:"cpu"`            // --cpu
	Memory         string            `yaml:"memory"`         // --memory
	Envs           map[string]string `yaml:"envs"`           // --envs
	Command        string            `yaml:"command"`        // --command
	Replicas       int               `yaml:"replicas"`       // --replicas
	Port           int               `yaml:"port"`           // --port
	RestfulPort    int               `yaml:"rest_api_port"`  // --restfulPort
	ModelName      string            `yaml:"modelName"`      // --modelName
	ModelPath      string            `yaml:"modelPath"`      // --modelPath
	EnableIstio    bool              `yaml:"enableIstio"`    // --enableIstio
	ServingName    string            `yaml:"servingName"`    // --servingName
	ServingVersion string            `yaml:"servingVersion"` // --servingVersion
	ModelDirs      map[string]string `yaml:"modelDirs"`
}

func (s ServeArgs) validateIstioEnablement() error {
	log.Debugf("--enableIstio=%t is specified.", s.EnableIstio)
	if !s.EnableIstio {
		return nil
	}
	
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.ServingName)
	if !matched {
		return fmt.Errorf("--serviceName should be numbers, letters, dashes, and underscores ONLY")
	}
	log.Debugf("--serviceVersion=%s is specified.", s.ServingVersion)
	if s.ServingVersion == "" {
		return fmt.Errorf("--serviceVersion must be specified if enableIstio=true")
	}
	
	return nil
}

func (s ServeArgs) validateModelName() error {
	if s.ModelName == "" {
		return fmt.Errorf("--modelName cannot be blank")
	}
	
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.ModelName)
	if !matched {
		return fmt.Errorf("--modelName should be numbers, letters, dashes, and underscores ONLY")
	}
	
	return nil
}

func ParseMountPath(dataset []string) (err error) {
	err = validate.ValidateDatasets(dataset)
	return err
}

func (serveArgs *ServeArgs) addServeCommonFlags(command *cobra.Command) {
	
	// create subcommands
	command.Flags().StringVar(&serveArgs.Image, "image", defaultTfServingImage, "the docker image name of serve job, default image is "+defaultTfServingImage)
	command.Flags().StringVar(&serveArgs.Command, "command", "", "the command will inject to container's command.")
	command.Flags().IntVar(&serveArgs.Gpus, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Cpu, "cpu", "", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Memory, "memory", "", "the request memory of each replica to run the serve.")
	command.Flags().IntVar(&serveArgs.Replicas, "replicas", 1, "the replicas number of the serve job.")
	command.Flags().StringVar(&serveArgs.ModelPath, "modelPath", "", "the model path for serving in the container")
	command.Flags().StringArrayVarP(&envs, "envs", "e", []string{}, "the environment variables")
	command.Flags().StringVar(&serveArgs.ModelName, "modelName", "", "the model name for serving")
	command.Flags().BoolVar(&serveArgs.EnableIstio, "enableIstio", false, "enable Istio for serving or not (disable Istio by default)")
	command.Flags().StringVar(&serveArgs.ServingName, "servingName", "", "the serving name")
	command.Flags().StringVar(&serveArgs.ServingVersion, "servingVersion", "", "the serving version")
	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>")
	
	command.MarkFlagRequired("servingName")
	
}

func init() {
	if os.Getenv(CHART_PKG_LOC) != "" {
		standalone_training_chart = filepath.Join(os.Getenv(CHART_PKG_LOC), "training")
	}
}

var (
	serveLong = `serve a job.

Available Commands:
  tensorflow,tf  Create a TensorFlow Serving Job.
    `
)

func NewServeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "serve",
		Short: "Serve a job.",
		Long:  serveLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewServingTensorFlowCommand())
	command.AddCommand(NewServingListCommand())
	command.AddCommand(NewServingDeleteCommand())
	command.AddCommand(NewTrafficRouterSplitCommand())
	return command
}
