package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/unisound-ail/atlasctl/cli"
	"github.com/unisound-ail/atlasctl/util"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/apimachinery/pkg/util/sets"

	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"strings"
)

var (
	showDetails bool
)

func NewTopNodeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "node",
		Short: "Display Resource (GPU) usage of nodes.",
		Run: func(cmd *cobra.Command, args []string) {
			client, _, err := cli.GetCliSetNameSpace()
			util.Must(err)
			allPods, err := acquireAllActivePods(client)
			util.MustE(err)

			nd := newNodeDescriber(client, allPods)
			nodeInfos, err := nd.getAllNodeInfos()
			util.MustE(err)

			displayTopNode(nodeInfos)
		},
	}
	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}


type NodeDescriber struct {
	client  *kubernetes.Clientset
	allPods []v1.Pod
}

type NodeInfo struct {
	node v1.Node
	pods []v1.Pod
}

func newNodeDescriber(client *kubernetes.Clientset, pods []v1.Pod) *NodeDescriber {
	return &NodeDescriber{
		client:  client,
		allPods: pods,
	}
}

func (nd *NodeDescriber) getAllNodeInfos() ([]NodeInfo, error) {
	nodeInfoList := []NodeInfo{}

	nodeList, err := nd.client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return nodeInfoList, err
	}

	for _, node := range nodeList.Items {
		pods := nd.getPodsFromNode(node)
		nodeInfoList = append(nodeInfoList, NodeInfo{
			node: node,
			pods: pods,
		})
	}

	return nodeInfoList, nil
}

func (nd *NodeDescriber) getPodsFromNode(node v1.Node) []v1.Pod {
	pods := []v1.Pod{}
	for _, pod := range nd.allPods {
		if pod.Spec.NodeName == node.Name {
			pods = append(pods, pod)
		}
	}
	return pods
}

func displayTopNode(nodes []NodeInfo) {
	if showDetails {
		displayTopNodeDetails(nodes)
	} else {
		displayTopNodeSummary(nodes)
	}
}

func displayTopNodeSummary(nodeInfos []NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUsInCluster     int64
		allocatedGPUsInCluster int64
	)

	fmt.Fprintf(w, "NAME\tIPADDRESS\tROLE\tGPU(Total)\tGPU(Allocated)\n")
	for _, nodeInfo := range nodeInfos {
		totalGPU, allocatedGPU := calculateNodeGPU(nodeInfo)
		totalGPUsInCluster += totalGPU
		allocatedGPUsInCluster += allocatedGPU

		address := "unknown"
		if len(nodeInfo.node.Status.Addresses) > 0 {
			// address = nodeInfo.node.Status.Addresses[0].Address
			for _, addr := range nodeInfo.node.Status.Addresses {
				if addr.Type == v1.NodeInternalIP {
					address = addr.Address
					break
				}
			}
		}

		/*role := "worker"
		if isMasterNode(nodeInfo.node) {
			role = "master"
		}*/
		
		role := strings.Join(findNodeRoles(&nodeInfo.node), ",")
		if len(role) == 0 {
			role = "<none>"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", nodeInfo.node.Name,
			address,
			role,
			strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatedGPU, 10))
	}
	fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\n")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
		strconv.FormatInt(allocatedGPUsInCluster, 10))
	var gpuUsage float64 = 0
	if totalGPUsInCluster > 0 {
		gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(allocatedGPUsInCluster, 10),
		strconv.FormatInt(totalGPUsInCluster, 10),
		int64(gpuUsage))
	// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)

	_ = w.Flush()
}

func displayTopNodeDetails(nodeInfos []NodeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalGPUsInCluster     int64
		allocatedGPUsInCluster int64
	)

	fmt.Fprintf(w, "\n")
	for _, nodeInfo := range nodeInfos {
		totalGPU, allocatedGPU := calculateNodeGPU(nodeInfo)
		totalGPUsInCluster += totalGPU
		allocatedGPUsInCluster += allocatedGPU

		address := "unknown"
		if len(nodeInfo.node.Status.Addresses) > 0 {
			address = nodeInfo.node.Status.Addresses[0].Address
		}
		
		role := strings.Join(findNodeRoles(&nodeInfo.node), ",")
		if len(role) == 0 {
			role = "<none>"
		}
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "NAME:\t%s\n", nodeInfo.node.Name)
		fmt.Fprintf(w, "IPADDRESS:\t%s\n", address)
		fmt.Fprintf(w, "ROLE:\t%s\n", role)

		pods := gpuPods(nodeInfo.pods)
		if len(pods) > 0 {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "NAMESPACE\tNAME\tGPU REQUESTS\tGPU LIMITS\n")
			for _, pod := range pods {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pod.Namespace,
					pod.Name,
					strconv.FormatInt(gpuInPod(pod), 10),
					strconv.FormatInt(gpuInPod(pod), 10))
			}
			fmt.Fprintf(w, "\n")
		}
		var gpuUsageInNode float64 = 0
		if totalGPU > 0 {
			gpuUsageInNode = float64(allocatedGPU) / float64(totalGPU) * 100
		} else {
			fmt.Fprintf(w, "\n")
		}

		fmt.Fprintf(w, "Total GPUs In Node %s:\t%s \t\n", nodeInfo.node.Name, strconv.FormatInt(totalGPU, 10))
		fmt.Fprintf(w, "Allocated GPUs In Node %s:\t%s (%d%%)\t\n", nodeInfo.node.Name, strconv.FormatInt(allocatedGPU, 10), int64(gpuUsageInNode))
		log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPU, 10),
			strconv.FormatInt(allocatedGPU, 10))

		fmt.Fprintf(w, "-----------------------------------------------------------------------------------------\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Allocated/Total GPUs In Cluster:\t")
	log.Debugf("gpu: %s, allocated GPUs %s", strconv.FormatInt(totalGPUsInCluster, 10),
		strconv.FormatInt(allocatedGPUsInCluster, 10))

	var gpuUsage float64 = 0
	if totalGPUsInCluster > 0 {
		gpuUsage = float64(allocatedGPUsInCluster) / float64(totalGPUsInCluster) * 100
	}
	fmt.Fprintf(w, "%s/%s (%d%%)\t\n",
		strconv.FormatInt(allocatedGPUsInCluster, 10),
		strconv.FormatInt(totalGPUsInCluster, 10),
		int64(gpuUsage))
	// fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ...)

	_ = w.Flush()
}

// findNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func findNodeRoles(node *v1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}
		
		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}

