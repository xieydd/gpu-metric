package cmd

import "k8s.io/api/core/v1"

// filter out the pods with GPU
func gpuPods(pods []v1.Pod) (podsWithGPU []v1.Pod) {
	for _, pod := range pods {
		if gpuInPod(pod) > 0 {
			podsWithGPU = append(podsWithGPU, pod)
		}
	}
	return podsWithGPU
}

// calculate the GPU count of each node
func calculateNodeGPU(nodeInfo NodeInfo) (totalGPU int64, allocatedGPU int64) {
	node := nodeInfo.node
	totalGPU = gpuInNode(node)
	// allocatedGPU = gpuInPod()

	for _, pod := range nodeInfo.pods {
		allocatedGPU += gpuInPod(pod)
	}

	return totalGPU, allocatedGPU
}

func isMasterNode(node v1.Node) bool {
	if _, ok := node.Labels[masterLabelRole]; ok {
		return true
	}

	return false
}

// The way to get GPU Count of Node: nvidia.com/gpu
func gpuInNode(node v1.Node) int64 {
	val, ok := node.Status.Capacity[NVIDIAGPUResourceName]

	if !ok {
		return gpuInNodeDeprecated(node)
	}

	return val.Value()
}

// The way to get GPU Count of Node: alpha.kubernetes.io/nvidia-gpu
func gpuInNodeDeprecated(node v1.Node) int64 {
	val, ok := node.Status.Capacity[DeprecatedNVIDIAGPUResourceName]

	if !ok {
		return 0
	}

	return val.Value()
}

func gpuInPod(pod v1.Pod) (gpuCount int64) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		gpuCount += gpuInContainer(container)
	}

	return gpuCount
}

func gpuInContainer(container v1.Container) int64 {
	val, ok := container.Resources.Limits[NVIDIAGPUResourceName]

	if !ok {
		return gpuInContainerDeprecated(container)
	}

	return val.Value()
}

func gpuInContainerDeprecated(container v1.Container) int64 {
	val, ok := container.Resources.Limits[DeprecatedNVIDIAGPUResourceName]

	if !ok {
		return 0
	}

	return val.Value()
}

func gpuInActivePod(pod v1.Pod) (gpuCount int64) {
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return 0
	}

	containers := pod.Spec.Containers
	for _, container := range containers {
		gpuCount += gpuInContainer(container)
	}

	return gpuCount
}
