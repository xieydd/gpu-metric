package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	//"github.com/unisound-ail/atlasctl/cli"
)

var helmCmd = []string{"helm"}
var KubeConfig string

/**
* install the release with cmd: helm install -f values.yaml chart_name
 */
func InstallRelease(name string, namespace string, values interface{}, chartName string) error {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	// 1. generate the template file
	valueFile, err := ioutil.TempFile(os.TempDir(), "values")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", valueFile.Name(), err)
		return err
	} else {
		log.Debugf("Save the values file %s", valueFile.Name())
	}
	// defer os.Remove(valueFile.Name())

	// 2. dump the object into the template file
	err = toYaml(values, valueFile)
	if err != nil {
		return err
	}

	// 3. check if the chart file exists, if it's it's unix path, then check if it's exist
	if strings.HasPrefix(chartName, "/") {
		if _, err = os.Stat(chartName); os.IsNotExist(err) {
			// TODO: the chart will be put inside the binary in future
			return err
		}
	}

	// 4. prepare the arguments
	args := []string{"install", "-f", valueFile.Name(), "--namespace", namespace, "--name", name, chartName}
	fmt.Printf("%s",args)
	log.Debugf("Exec %s, %v", binary, args)
	
	env := os.Environ()
	if KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	fmt.Println("")
	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Fatalf("Failed to execute %s, %v with %v", binary, args, err)
	}

	// 6. clean up the value file if needed
	if log.GetLevel() != log.DebugLevel {
		err = os.Remove(valueFile.Name())
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", valueFile.Name(), err)
		}
	}

	return nil
}

/**
* check if the release exist
 */
func CheckRelease(name string) (exist bool, err error) {
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return exist, err
	}

	cmd := exec.Command(helmCmd[0], "get", name)
	// support multiple cluster management
	if KubeConfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v", err)
		return exist, err
	}

	err = cmd.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus := status.ExitStatus()
				log.Debugf("Exit Status: %d", exitStatus)
				if exitStatus == 1 {
					err = nil
				}
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
			return exist, err
		}
	} else {
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if waitStatus.ExitStatus() == 0 {
			exist = true
		} else {
			if waitStatus.ExitStatus() != -1 {
				return exist, fmt.Errorf("unexpected return code %d when exec helm get %s", waitStatus.ExitStatus(), name)
			}
		}
	}

	return exist, err
}

func DeleteRelease(name string) error {
	cmd, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	args := []string{"helm", "del", "--purge", name}

	env := os.Environ()
	if KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}
	return syscall.Exec(cmd, args, env)
}

func ListReleases() (releases []string, err error) {
	releases = []string{}
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return releases, err
	}

	cmd := exec.Command(helmCmd[0], "list", "-q")
	// support multiple cluster management
	if KubeConfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}
	out, err := cmd.Output()
	if err != nil {
		return releases, err
	}
	return strings.Split(string(out), "\n"), nil
}

func ListReleaseMap() (releaseMap map[string]string, err error) {
	releaseMap = map[string]string{}
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return releaseMap, err
	}

	cmd := exec.Command(helmCmd[0], "list")
	// support multiple cluster management
	if KubeConfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}
	out, err := cmd.Output()
	if err != nil {
		return releaseMap, err
	}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.Trim(line, " ")
		if !strings.Contains(line, "NAME") {
			cols := strings.Fields(line)
			log.Debugf("cols: ", cols, len(cols))
			if len(cols) > 1 {
				log.Debugf("releaseMap: %s=%s\n", cols[0], cols[len(cols)-1])
				releaseMap[cols[0]] = cols[len(cols)-1]
			}
		}
	}

	return releaseMap, nil
}

func ListAllReleasesWithDetail() (releaseMap map[string][]string, err error) {
	releaseMap = map[string][]string{}
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return releaseMap, err
	}
	/*_,namespace,err := cli.GetCliSetNameSpace()
	if err == nil {
		fmt.Printf("%s",err)
		os.Exit(0)
	}
	args := []string{"list","--namespace",namespace}*/
	
	cmd := exec.Command(helmCmd[0],"list","--all")
	// support multiple cluster management
	if KubeConfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}
	out, err := cmd.Output()
	if err != nil {
		return releaseMap, err
	}
	lines := strings.Split(string(out), "\n")
	
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if !strings.Contains(line, "NAME") {
			cols := strings.Fields(line)
			log.Debugf("cols: ", cols, len(cols))
			if len(cols) > 3 {
				log.Debugf("releaseMap: %s=%s\n", cols[0], cols)
				releaseMap[cols[0]] = cols
			}
		}
	}
	return releaseMap, err
}


func toYaml(values interface{}, file *os.File) error {
	log.Debugf("values: %+v", values)
	data, err := yaml.Marshal(values)
	if err != nil {
		log.Errorf("Failed to marshal value %v due to %v", values, err)
		return err
	}

	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		log.Errorf("Failed to write %v to %s due to %v", data, file.Name(), err)
	}
	return err
}
