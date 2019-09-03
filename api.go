package main

import (
	"os"
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
	Name             string `form:"name" binding:"Required"`
	ClusterIP        string
	ProxyURL         string
	Routes           []Route
	Version          string `form:"version"` // TODO: Implement different kind cluster versions
	Kubeconfig       string
	Federated        bool
	InstanceEndpoint string
}

func NewAPIResult(output string) APIResult {

	// Get running cluster in each successful response
	res, err := Kind("get", "clusters")
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

	return APIResult{LocalClusters: localClusters, AvailableClusters: clusterNames, Clusters: clusters, ActiveEndpoints: activeEndpoints, Output: output, ClusterIPs: clusterIPs}
}

func GetClusterInfo(clustername string) (KubernetesCluster, error) {
	var kubehost, host string

	kubehost = os.Getenv("KUBEHOST")
	host = os.Getenv("HOST")
	if len(kubehost) > 0 {
		host = kubehost
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
