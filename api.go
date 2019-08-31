package main

import (
	"github.com/pkg/errors"
	"os"
	"strings"
)

type APIResult struct {
	Output            string
	AvailableClusters []string
	Clusters          map[string]KubernetesCluster
	ActiveEndpoints   map[string]string
	ClusterIPs        map[string]string

	Error string
}

type KubernetesCluster struct {
	Name      string `form:"name" binding:"Required"`
	ClusterIP string
	ProxyURL  string
	Routes    []Route
	Version   string `form:"version"` // TODO: Implement different kind cluster versions
}

func NewAPIResult(output string) APIResult {

	// Get running cluster in each successful response
	res, err := Kind("get", "clusters")
	if err != nil {
		return APIResult{Error: err.Error(), Output: res}
	}
	clusters := make(map[string]KubernetesCluster)
	var clusterNames []string
	if len(res) > 0 {
		clusterNames = strings.Split(res, "\n")
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

	return APIResult{AvailableClusters: clusterNames, Clusters: clusters, ActiveEndpoints: activeEndpoints, Output: output, ClusterIPs: clusterIPs}
}

func GetClusterInfo(clustername string) (KubernetesCluster, error) {
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
	return KubernetesCluster{Name: clustername, ClusterIP: ip, ProxyURL: os.Getenv("HOST") + ":" + cluster.Port, Routes: routes}, nil
}
