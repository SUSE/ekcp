package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Bridge struct {
	IPAddress string `json:"IPAddress"`
}
type Networks struct {
	Bridge Bridge `json:"bridge"`
}
type NetworkSettings struct {
	Networks Networks `json:"Networks"`
}
type DockerInspect struct {
	NetworkSettings NetworkSettings `json:"NetworkSettings"`
}

func Kind(envs []string, args ...string) (string, error) {

	k := exec.Command("kind", args...)
	k.Env = os.Environ()
	for _, e := range envs {
		k.Env = append(k.Env, e)
	}
	out, err := k.CombinedOutput()
	output := string(out)
	output = strings.TrimSuffix(output, "\n")

	if err != nil {
		return output, err
	}
	return output, nil
}

func Docker(args ...string) (string, error) {
	out, err := exec.Command("docker", args...).CombinedOutput()
	output := string(out)
	output = strings.TrimSuffix(output, "\n")

	if err != nil {
		return output, err
	}
	return output, nil
}

func GetKubeIP(cluster string) (string, error) {
	j, err := Docker("inspect", cluster+"-control-plane")
	if err != nil {
		return j, err
	}
	v := []DockerInspect{}
	err = json.Unmarshal([]byte(j), &v)
	if err != nil {
		return j, err
	}
	if len(v) != 1 {
		return "", errors.New("couldn't decode Kind cluster IP")
	}
	return v[0].NetworkSettings.Networks.Bridge.IPAddress, nil
}
