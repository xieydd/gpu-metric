package create

import (
	"encoding/json"
	"fmt"
	"log"
	"os/user"
	strconv2 "strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/unisound-ail/atlasctl/err"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/spf13/cobra"
	"github.com/unisound-ail/atlasctl/cli"
	"github.com/unisound-ail/atlasctl/util"
)

// Opts is params
type Opts struct {
	Name       string
	Image      string
	Arguments  string
	WorkingDir string
	CPU        string
	Memory     string
	GPU        int64
	Volumes    []string
	NodeName   string
	AddUser    bool
	NodeLabels []string
	DryRun     bool
	PullAlways bool
	Envs       []string
}

// Volumeinfo is for docker -v param
type Volumeinfo struct {
	Type     string
	IP       string
	SrcPath  string
	DestPath string
}

//Run is main for create
func Run() {

	clientset, namespace, e := cli.GetCliSetNameSpace()
	util.Must(e)

	e = create(clientset, namespace)
	util.Must(e)
}

var opts = Opts{}
func  AddCommonFlags(command *cobra.Command) {
	command.Flags().BoolVar(&opts.AddUser, "adduser", false, "add User in containter")
	command.Flags().StringVarP(&opts.Name, "name", "n", "", "task name")
	command.Flags().StringVarP(&opts.Image, "image", "i", "", "docker image name")
	command.Flags().StringVarP(&opts.Arguments, "args", "a", "", "default command")
	command.Flags().StringVarP(&opts.WorkingDir, "workingdir", "w", "", "workingdir")
	command.Flags().StringVarP(&opts.CPU, "cpu", "c", "", "CPU quality")
	command.Flags().StringVarP(&opts.Memory, "mem", "m", "", "memory quality")
	command.Flags().Int64VarP(&opts.GPU, "gpu", "g", 0, "GPU quality")
	command.Flags().StringSliceVarP(&opts.Volumes, "volumes", "v", nil, "volumes")
	command.Flags().StringVarP(&opts.NodeName, "nodename", "N", "", "node name")
	command.Flags().StringSliceVarP(&opts.NodeLabels, "nodeLabels", "l", nil, "node labels")
	command.Flags().BoolVar(&opts.DryRun, "dryrun", false, "dry run")
	command.Flags().BoolVar(&opts.PullAlways, "pull", false, "pull always")
	command.Flags().StringSliceVarP(&opts.Envs, "environment", "e", nil, "pass env to container")
}

func processVolume(volstrlist []string) ([]Volumeinfo, error) {
	var volumeinfo []Volumeinfo

	for len(volstrlist) > 0 {

		switch strings.ToLower(volstrlist[0]) {

		case "nfs":
			ippath := strings.Split(volstrlist[1], ":")
			if len(ippath) != 2 {
				return volumeinfo, err.NewError(err.ErrInvalidVolume, "invaild volume format")

			}

			volumeinfo = append(volumeinfo, Volumeinfo{Type: "NFS", IP: ippath[0], SrcPath: ippath[1], DestPath: volstrlist[2]})

		case "hostpath":
			volumeinfo = append(volumeinfo, Volumeinfo{Type: "hostPath", SrcPath: volstrlist[1], DestPath: volstrlist[2]})

		default:
			return volumeinfo, err.NewError(err.ErrInvalidVolume, "invaild volume type")
		}

		volstrlist = volstrlist[3:]
	}

	return volumeinfo, nil

}

func formatEnv(envMap map[string]string) []v1.EnvVar {
	var envVar []v1.EnvVar
	for k, v := range envMap {
		envVar = append(envVar, v1.EnvVar{Name: k, Value: v})
	}
	return envVar
}

func processEnv(envStrList []string, m map[string]string) error {
	for _, envStr := range envStrList {
		env := strings.Split(envStr, "=")
		if len(env) != 2 {
			return err.NewError(err.ErrInvalidEnv, "invaild environment format")
		}
		m[env[0]] = env[1]
	}
	return nil
}

func create(clientset *kubernetes.Clientset,
	namespace string) error {

	var volumes []v1.Volume
	var volumemounts []v1.VolumeMount

	envMap := make(map[string]string)
	if e := processEnv(opts.Envs, envMap); e != nil {
		return e
	}

	volumelist, e := processVolume(opts.Volumes)
	if e != nil {
		return e
	}

	for index, vol := range volumelist {
		name := strings.ToLower(vol.Type) + strconv2.Itoa(index)

		if vol.Type == "NFS" {
			volumes = append(volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{NFS: &v1.NFSVolumeSource{Server: vol.IP, Path: vol.SrcPath}}})
			volumemounts = append(volumemounts, v1.VolumeMount{Name: name, MountPath: vol.DestPath})

		} else if vol.Type == "hostPath" {
			volumes = append(volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: vol.SrcPath}}})
			volumemounts = append(volumemounts, v1.VolumeMount{Name: name, MountPath: vol.DestPath})

		}

	}

	lim := &v1.ResourceList{}
	req := &v1.ResourceList{}

	if opts.Memory != "" {
		parsed, e := resource.ParseQuantity(opts.Memory)
		if e != nil {
			return e

		}
		(*req)[v1.ResourceMemory] = parsed

	}

	if opts.CPU != "" {
		parsed, e := resource.ParseQuantity(opts.CPU)
		if e != nil {
			return e

		}
		(*req)[v1.ResourceCPU] = parsed

	}

	if opts.GPU > 0 {
		(*lim)["nvidia.com/gpu"] = *resource.NewQuantity(opts.GPU, resource.DecimalSI)
	} else {
		envMap["NVIDIA_VISIBLE_DEVICES"] = ""
	}

	args := opts.Arguments

	if opts.AddUser {
		userinfo, e := user.Current()
		if e != nil {
			log.Panic("Get Current User Error")
		}

		gids, e := userinfo.GroupIds()
		if e != nil {
			fmt.Println(e)
			log.Panic("Get Current User's GIDs Error")
		}
		group_cmd := ""
		group_ids := ""
		for _, gid := range gids {
			g, e := user.LookupGroupId(gid)
			if e != nil {
				fmt.Printf("%+v.GroupIds(): %v", g, e)
			} else {
				group_cmd = group_cmd + fmt.Sprintf("groupadd -f -g %s %s && ", g.Gid, g.Name)
				if gid != userinfo.Gid {
					group_ids = group_ids + "," + gid
				}
			}
		}
		tmp := []rune(group_ids)
		if len(tmp) > 0 {
			group_ids = "-G " + string(tmp[1:])
		}
		args = fmt.Sprintf("%s useradd -m -u %s -g %s %s -s /bin/bash %s; %s", group_cmd, userinfo.Uid, userinfo.Gid, group_ids, userinfo.Username, args)
	}

	context := v1.PodSecurityContext{}

	container := v1.Container{
		Name:            opts.Name,
		Image:           opts.Image,
		Command:         []string{"/bin/bash"},
		Args:            []string{"-c", args},
		ImagePullPolicy: v1.PullIfNotPresent,
		Resources:       v1.ResourceRequirements{Limits: *lim, Requests: *req},
		VolumeMounts:    volumemounts,
		Env:             formatEnv(envMap),
	}

	if opts.PullAlways {
		container.ImagePullPolicy = v1.PullAlways
	}

	if opts.WorkingDir != "" {
		container.WorkingDir = opts.WorkingDir
	}

	terminationgraceperiodseconds := int64(1)

	affinity := v1.Affinity{}
	if opts.NodeLabels != nil {
		affinity.NodeAffinity = &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "alpha.kubernetes.io/nvidia-gpu-name",
								Operator: v1.NodeSelectorOpIn,
								Values:   opts.NodeLabels,
							},
						},
					},
				},
			},
		}
	}

	pj := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: opts.Name,
			Labels: map[string]string{
				"name": opts.Name,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy:                 v1.RestartPolicyNever,
			DNSPolicy:                     v1.DNSClusterFirst,
			TerminationGracePeriodSeconds: &terminationgraceperiodseconds,
			Containers:                    []v1.Container{container},
			Volumes:                       volumes,
			NodeName:                      opts.NodeName,
			Affinity:                      &affinity,
			SecurityContext:               &context,
		},
	}

	//print yaml
	bb2, err4 := json.Marshal(pj)
	if err4 != nil {
		fmt.Println(err4)
	}

	jj, err2 := yaml.JSONToYAML(bb2)
	if err2 != nil {
		return err2
	}
	fmt.Println(string(jj))

	//run
	if !opts.DryRun {
		p2, err3 := clientset.CoreV1().Pods(namespace).Create(&pj)
		if err3 != nil {
			return err3

		}
		fmt.Printf("Pod %s have created success\n", p2.Name)

	}

	return nil
}
