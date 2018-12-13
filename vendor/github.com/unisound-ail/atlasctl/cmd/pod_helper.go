package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"github.com/unisound-ail/atlasctl/util"
)

// acquire all active pods from all namespaces
func acquireAllActivePods(client *kubernetes.Clientset) ([]v1.Pod, error) {
	allPods := []v1.Pod{}

	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	util.MustE(err)
	nodeNonTerminatedPodsList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	util.MustE(err)

	for _, pod := range nodeNonTerminatedPodsList.Items {
		allPods = append(allPods, pod)
	}
	return allPods, nil
}

func acquireAllPods(client *kubernetes.Clientset) ([]v1.Pod, error) {
	allPods := []v1.Pod{}
	podList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return allPods, err
	}
	for _, pod := range podList.Items {
		allPods = append(allPods, pod)
	}
	return allPods, nil
}

func acquireAllJobs(client *kubernetes.Clientset) ([]batchv1.Job, error) {
	allJobs := []batchv1.Job{}
	jobList, err := client.BatchV1().Jobs(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return allJobs, err
	}
	for _, job := range jobList.Items {
		allJobs = append(allJobs, job)
	}
	return allJobs, nil
}

func getPodNameFromJob(client kubernetes.Interface, namespace, name string) (podName string, err error) {
	pods, err := jobPods(client, namespace, name)
	if err != nil {
		return "", err
	}

	if len(pods) == 0 {
		return "", fmt.Errorf("Failed to find the pod for job %s, maybe you need to set --namespace", name)
	}

	for _, pod := range pods {
		meta := pod.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				break
			}
		}

		if isJob {
			return pod.Name, nil
		}
	}

	return "", fmt.Errorf("getPodNameFromJob: Failed to find the pod of job")
}

// Get the latest pod from the Kubernetes job
func getPodFromJob(client kubernetes.Interface, namespace, releaseName string) (jobPod v1.Pod, err error) {
	pods, err := jobPods(client, namespace, releaseName)
	if err != nil {
		return jobPod, err
	}

	if len(pods) == 0 {
		return jobPod, fmt.Errorf("getPodFromJob: Failed to find the pod for job %s, maybe you need to set --namespace", name)
	}

	var latest metav1.Time

	for _, pod := range pods {
		meta := pod.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				break
			}
		}

		if isJob {
			// return pod, nil
			if jobPod.Name == "" {
				latest = pod.CreationTimestamp
				jobPod = pod
				log.Debugf("set pod %s as first jobpod, and it's time is %v", jobPod.Name, jobPod.CreationTimestamp)
			} else {
				log.Debugf("current jobpod %s , and it's time is %v", jobPod.Name, latest)
				log.Debugf("candidate jobpod %s , and it's time is %v", pod.Name, pod.CreationTimestamp)
				current := pod.CreationTimestamp
				if latest.Before(&current) {
					jobPod = pod
					latest = current
					log.Debugf("replace")
				} else {
					log.Debugf("no replace")
				}
			}
		}
	}

	if jobPod.Name == "" {
		err = fmt.Errorf("Not able to job with release %s in pods %v", releaseName, pods)
	}

	return jobPod, err
}

// List all the pods which associate to the atlasctl jobs, including the pods in the statefulset and the job
func listAllPodsForJob(client kubernetes.Interface, namespace string, releaseName string) (pods []v1.Pod, err error) {
	podList, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", releaseName),
	})

	if err != nil {
		return nil, err
	}

	pods = []v1.Pod{}

	for _, item := range podList.Items {
		meta := item.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				log.Debugf("find job pod %v, break", item)
				break
			}
		}

		if !isJob {
			pods = append(pods, item)
			log.Debugf("add pod %v to pods", item)
		}
	}

	jobPod, err := getPodFromJob(client, namespace, releaseName)
	if err != nil {
		return nil, err
	}

	pods = append(pods, jobPod)

	return pods, err
}

func jobPods(client kubernetes.Interface, namespace string, releaseName string) ([]v1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", releaseName),
	})

	if err != nil {
		return nil, err
	}

	return podList.Items, err
}
