package main

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type APIResult struct {
	Output            string
	AvailableClusters []string
	Clusters          map[string]KubernetesCluster
	ActiveEndpoints   map[string]string
	ClusterIPs        map[string]string
	LocalClusters     []string

	Error string
}

type KubernetesCluster struct {
	Name                 string `form:"name" binding:"Required"`
	ClusterIP            string
	ProxyURL             string
	Routes               []Route
	NodeImage            string `form:"node_image"`
	Version              string `form:"version"` // TODO: Implement different kind versions
	Kubeconfig           string
	Federated            bool
	InstanceEndpoint     string
	RawEncodedKindConfig string `form:"kindconfig"` // base64 encoded config file
}

func (kc *KubernetesCluster) HasConfig() bool {
	return len(kc.RawEncodedKindConfig) > 0
}

func (kc *KubernetesCluster) HasNodeImage() bool {
	return len(kc.NodeImage) > 0
}

func (kc *KubernetesCluster) DecodeConfig() ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(kc.RawEncodedKindConfig)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (kc *KubernetesCluster) DecodeKubeConfig() ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(kc.Kubeconfig)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (kc *KubernetesCluster) WriteConfig(path string) error {
	config, err := kc.DecodeConfig()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, config, os.ModePerm)
}

func (kc *KubernetesCluster) Start() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "Could not find user home")
	}
	path := filepath.Join(usr.HomeDir, ".kube", "kind-config-"+kc.Name)
	env := []string{"KUBECONFIG=" + path}

	args := []string{"create", "cluster", "--name", kc.Name}
	if kc.HasConfig() {
		tmpFile, err := ioutil.TempFile(os.TempDir(), "ekcp-")
		if err != nil {
			return "", err
		}

		// Remember to clean up the file afterwards
		defer os.Remove(tmpFile.Name())

		if err := kc.WriteConfig(tmpFile.Name()); err != nil {
			return "", err
		}

		args = append(args, "--config", tmpFile.Name())
	}
	if kc.HasNodeImage() {
		args = append(args, "--image", kc.NodeImage)
	}

	return Kind(env, args...)
}

func NewAPIResult(output string) APIResult {

	// Get running cluster in each successful response
	res, err := Kind([]string{}, "get", "clusters")
	if err != nil {
		return APIResult{Error: err.Error(), Output: res}
	}
	clusters := make(map[string]KubernetesCluster)
	var clusterNames []string
	var localClusters []string

	if len(res) > 0 {
		clusterNames = strings.Split(res, "\n")
		localClusters = clusterNames
	}

	activeEndpoints := make(map[string]string)
	clusterIPs := make(map[string]string)
	// Get active endpoints
	for _, cluster := range clusterNames {
		kCluster, err := GetClusterInfo(cluster)
		if err != nil {
			continue
		}
		activeEndpoints[cluster] = kCluster.ProxyURL
		clusterIPs[cluster] = kCluster.ClusterIP
		clusters[cluster] = kCluster
	}

	if Federation.HasSlaves() {
		kubeClusters := Federation.List()
		for _, kubeC := range kubeClusters {
			clusterNames = append(clusterNames, kubeC.Name)
			clusterIPs[kubeC.Name] = kubeC.ClusterIP
			clusters[kubeC.Name] = kubeC
		}
	}

	if len(Proxied.ExternalClusters()) != 0 {
		for _, kubeC := range Proxied.ExternalClusters() {
			_, err := Proxied.GetKubeConfig(kubeC)
			if err != nil {
				continue
			}
			kc := KubernetesCluster{
				Name:       kubeC,
				Kubeconfig: "http://" + os.Getenv("KUBEHOST") + ":" + os.Getenv("PORT") + "/api/v1/cluster/" + kubeC + "/kubeconfig", // To retreive it, use Proxied.GetKubeConfig
			}
			clusterNames = append(clusterNames, kubeC)
			clusterIPs[kubeC] = ""
			clusters[kubeC] = kc
		}
	}

	return APIResult{LocalClusters: localClusters, AvailableClusters: clusterNames, Clusters: clusters, ActiveEndpoints: activeEndpoints, Output: output, ClusterIPs: clusterIPs}
}

func GetClusterInfo(clustername string) (KubernetesCluster, error) {
	var kubehost, host string

	kubehost = os.Getenv("KUBEHOST")
	host = os.Getenv("HOST")
	if len(kubehost) > 0 {
		host = kubehost
	}

	_, err := Proxied.GetKubeConfig(clustername)
	if err == nil {
		return KubernetesCluster{
			Name:       clustername,
			Kubeconfig: "http://" + os.Getenv("KUBEHOST") + ":" + os.Getenv("PORT") + "/api/v1/cluster/" + clustername + "/kubeconfig",
		}, nil
	}

	cluster, ok := Proxied.Endpoints[clustername]
	if !ok {
		return KubernetesCluster{}, errors.New("Cluster not proxied")
	}
	ip, err := GetKubeIP(clustername)
	if err != nil {
		return KubernetesCluster{}, errors.Wrap(err, "Failed to get cluster ip")
	}
	rr, err := NewRouteRegister()
	if err != nil {
		return KubernetesCluster{}, errors.Wrap(err, "Failed to get route register ip")
	}
	routes, err := rr.ClusterRoutes(clustername)
	if err != nil {
		return KubernetesCluster{}, errors.Wrap(err, "Failed to get cluster routes")
	}
	return KubernetesCluster{Kubeconfig: "http://" + os.Getenv("KUBEHOST") + ":" + os.Getenv("PORT") + "/api/v1/cluster/" + clustername + "/kubeconfig", Name: clustername, ClusterIP: ip, ProxyURL: host + ":" + cluster.Port, Routes: routes}, nil
}
