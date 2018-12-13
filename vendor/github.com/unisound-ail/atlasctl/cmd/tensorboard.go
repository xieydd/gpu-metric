package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/unisound-ail/atlasctl/util"
	"github.com/unisound-ail/atlasctl/cli"
)

type createTensorboardArgs struct {
	UseTensorboard   bool   `yaml:"useTensorboard"`   // --tensorboard
	TensorboardImage string `yaml:"tensorboardImage"` // --tensorboardImage
	TrainingLogdir   string `yaml:"trainingLogdir"`   // --logdir
	HostLogPath      string `yaml:"hostLogPath"`
}

func tensorboardURL(name, namespace string) (url string, err error) {

	// 1. Get address
	clientset, namespace, e := cli.GetCliSetNameSpace()
	util.MustE(e)
	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	node := v1.Node{}
	findReadyNode := false

	for _, item := range nodeList.Items {
		for _, condition := range item.Status.Conditions {
			if condition.Type == "Ready" {
				if condition.Status == "True" {
					node = item
					findReadyNode = true
					break
				}
			}
		}
	}

	if !findReadyNode {
		return "", fmt.Errorf("Failed to find the ready node for exporting tensorboard.")
	}

	address := node.Status.Addresses[0].Address

	// 2. Get port

	serviceList, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,role=tensorboard", name),
	})
	if err != nil {
		// if errors.IsNotFound(err) {
		// 	log.Debugf("The tensorboard service doesn't exist")
		// 	return "", nil
		// }else{
		// 	return "", err
		// }
		return "", err
	}

	if len(serviceList.Items) == 0 {
		log.Debugf("Failed to find the tensorboard service due to service"+
			"List is empty when selector is release=%s,role=tensorboard.", name)
		return "", nil
	}

	ports := serviceList.Items[0].Spec.Ports
	if len(ports) == 0 {
		log.Debugf("Failed to find the tensorboard service due to ports list is empty.")
		return "", nil
	}

	nodePort := ports[0].NodePort
	url = fmt.Sprintf("%s:%d", address, nodePort)

	return url, nil
}
