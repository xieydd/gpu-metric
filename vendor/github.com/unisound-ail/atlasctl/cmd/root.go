// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

var cfgFile string
var version bool
var verStr string

var (
	loadingRules *clientcmd.ClientConfigLoadingRules
	logLevel     string
	enablePProf  bool
)

const (
	CLIName = "atlasctl"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "atlasctl",
	Short: "atlasctl is the utility tool for ATLAS project",
	Long:  `atlasctl is the utility tool for ATLAS project`,
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Println(verStr)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(versionstr string) {
	verStr = versionstr
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)
	
	addKubectlFlagsToCmd(RootCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.atlasctl.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolVarP(&version, "version", "", false, "Show version")
	
	// enable logging
	RootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")
	RootCmd.PersistentFlags().BoolVar(&enablePProf, "pprof", false, "enable cpu profile")
	RootCmd.PersistentFlags().StringVar(&atlasNamespace, "atlasNamespace", "atlas-system", "The namespace of atlas system service, like TFJob")

	RootCmd.AddCommand(NewCreateCommand())
	RootCmd.AddCommand(NewServeCommand())
	RootCmd.AddCommand(NewTopCommand())
	RootCmd.AddCommand(NewVersionCmd(CLIName))
	RootCmd.AddCommand(NewGetCommand())
	RootCmd.AddCommand(NewListCommand())
	RootCmd.AddCommand(NewDeleteCommand())
	RootCmd.AddCommand(NewLogsCommand())
	RootCmd.AddCommand(NewLogViewerCommand())
	RootCmd.AddCommand(NewDataCommand())
	return RootCmd
}

func addKubectlFlagsToCmd(cmd *cobra.Command) {
	// The "usual" clientcmd/kubectl flags
	loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	// kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "config", "", "Path to a kube config. Only required if out-of-cluster")
	//cmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "the namespace of the job")
	// clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}


// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".atlasctl2" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".atlasctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}


func createNamespace(client *kubernetes.Clientset, namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.Core().Namespaces().Create(ns)
	return err
}

func getNamespace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return client.Core().Namespaces().Get(namespace, metav1.GetOptions{})
}


func ensureNamespace(client *kubernetes.Clientset, namespace string) error {
	_, err := getNamespace(client, namespace)
	if err != nil && errors.IsNotFound(err) {
		return createNamespace(client, namespace)
	}
	return err
}
