package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type ClusterType int

const (
	DockerCluster  = 0
	CrioCluster    = iota
	UnknownCluster = iota
)

type CrioImages struct {
	RepoTags []string `json:"repoTags"`
}

type CrioList struct {
	Images []CrioImages `json:"images"`
}

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

func GetKubeBackend(cluster string) (ClusterType, error) {
	_, err := Docker("exec", cluster+"-control-plane", "crictl", "--help")
	if err == nil {
		return ClusterType(CrioCluster), nil
	}

	_, err = Docker("exec", cluster+"-control-plane", "docker", "--help")
	if err == nil {
		return ClusterType(DockerCluster), nil
	}

	return ClusterType(UnknownCluster), errors.New("Could not get the cluster backend type")
}

func GetClusterImages(cluster string) ([]string, error) {
	clusterType, err := GetKubeBackend(cluster)

	if err != nil {
		return []string{}, errors.Wrap(err, "failed getting kubernetes backend type")
	}

	switch clusterType {
	case CrioCluster:
		j, err := Docker("exec", cluster+"-control-plane", "crictl", "images", "-o", "json")
		if err != nil {
			return []string{}, errors.Wrap(err, "failed executing crictl images")
		}

		var crijson CrioList

		err = json.Unmarshal([]byte(j), &crijson)
		if err != nil {
			return []string{}, errors.Wrap(err, "failed parsing crictl images json")
		}
		var imageList []string
		for _, img := range crijson.Images {
			image := img.RepoTags[0]
			imageList = append(imageList, image)
		}

		return imageList, nil

	case DockerCluster:
		j, err := Docker("exec", cluster+"-control-plane", "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
		if err != nil {
			return []string{}, errors.Wrap(err, "failed executing docker images")
		}

		return strings.Split(j, "\n"), nil
	case UnknownCluster:
		return []string{}, errors.New("Cannot get image list from unknown cluster type")
	default:
		return []string{}, errors.New("Unreachable code")

	}

	//return []string{}, errors.New("Unreachable code")
}
