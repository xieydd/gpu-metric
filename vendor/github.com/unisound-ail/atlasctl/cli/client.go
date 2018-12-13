package cli

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"fmt"
)

// GetCliSetNameSpace get namespaced clientset
func GetCliSetNameSpace() (*kubernetes.Clientset, string, error) {

	// uses the current context in ~/.kube/config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	namespace, override, err := kubeConfig.Namespace()

	if override == true {
		glog.Info("namespace is override!")

	}

	if err != nil {
		return nil, "", err

	}

	//spew.Dump(namespace, override)

	config, err := kubeConfig.ClientConfig()
	
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", err

	}

	fmt.Errorf("%s",err)

	return clientset, namespace, nil

}
