package main

import (
	"os"
	"strings"
)

type APIResult struct {
	Output          string
	Clusters        []string
	ActiveEndpoints map[string]string
	ClusterIPs      map[string]string

	Error string
}

func NewAPIResult(output string) APIResult {

	// Get running cluster in each successful response
	res, err := Kind("get", "clusters")
	if err != nil {
		return APIResult{Error: err.Error(), Output: res}
	}
	var clusters []string

	if len(res) > 0 {
		clusters = strings.Split(res, "\n")
	}

	activeEndpoints := make(map[string]string)
	clusterIPs := make(map[string]string)
	// Get active endpoints
	for cluster, p := range Proxied.Endpoints {
		activeEndpoints[cluster] = os.Getenv("HOST") + ":" + p.Port
		ip, _ := GetKubeIP(cluster)
		clusterIPs[cluster] = ip
	}

	return APIResult{Clusters: clusters, ActiveEndpoints: activeEndpoints, Output: output, ClusterIPs: clusterIPs}
}
