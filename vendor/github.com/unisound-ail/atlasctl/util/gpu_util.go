package util
//
//import (
//	"k8s.io/api/core/v1"
//	"k8s.io/kubernetes/test/e2e/framework"
//	. "github.com/onsi/gomega"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/util/uuid"
//)
//
//const (
//	NVIDIAGPUResourceName = "nvidia.com/gpu"
//	GPUDevicePluginDSYAML = "https://raw.githubusercontent.com/kubernetes/kubernetes/master/cluster/addons/device-plugins/nvidia-gpu/daemonset.yaml"
//)
//
//// NumGPU in one Node
//func NumberOFNVIDIAGPUs(node v1.Node) int64 {
//	val, ok := node.Status.Capacity[NVIDIAGPUResourceName]
//	if !ok {
//		return 0
//	}
//
//	return val.Value()
//}
//
////Create NVIDIADevicePlugin Pod Base on a office daemonset template spec
//func NVIDIADevicePlugin(ns string) *v1.Pod {
//	ds, err := framework.DsFromManifest(GPUDevicePluginDSYAML)
//	Expect(err).NotTo(HaveOccurred())
//	p := &v1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "device-plugin-nvidia-gpu-" + string(uuid.NewUUID()),
//			Namespace: ns,
//		},
//
//		Spec: ds.Spec.Template.Spec,
//	}
//	// Remove node affinity
//	p.Spec.Affinity = nil
//
//	return p
//}
//
//func GetGPUDevicePluginImage() string {
//	ds, err := framework.DsFromManifest(GPUDevicePluginDSYAML)
//	if err != nil || ds == nil || len(ds.Spec.Template.Spec.Containers) < 1 {
//		return ""
//	}
//	return ds.Spec.Template.Spec.Containers[0].Image
//}