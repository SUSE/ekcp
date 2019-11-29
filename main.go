package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/go-macaron/binding"
	"github.com/pkg/errors"
	macaron "gopkg.in/macaron.v1"
)

func main() {
	m := macaron.Classic()
	err := ProxyStartup()
	if err != nil {
		panic(err)
	}
	m.Use(macaron.Renderer())
	if os.Getenv("ROUTE_REGISTER") == "true" {
		m.Use(MacaronRR(os.Getenv("DOMAIN")))
	}
	m.Get("/:id", GetKubeConfig)
	m.Post("/new", binding.Bind(KubernetesCluster{}), NewCluster)
	m.Delete("/:id", DeleteCluster)
	m.Get("/", ListClusters)
	m.Get("/kubeconfig/:id", GetProxyKubeConfig)
	m.Get("/kube/:id", GetKubeEndpoint)

	m.Get("/api/v1/cluster/:id/info", ClusterInfo)
	m.Get("/api/v1/cluster/:id/kubeconfig", GetProxyKubeConfig)
	m.Get("/api/v1/cluster/:id/e2e/kubeconfig", GetKubeConfig)
	m.Post("/api/v1/cluster/insert", binding.Bind(KubernetesCluster{}), InsertCluster)

	m.Post("/api/v1/cluster/new", binding.Bind(KubernetesCluster{}), NewCluster)
	m.Get("/api/v1/cluster", ListClusters)
	m.Delete("/api/v1/cluster/:id", DeleteCluster)

	m.Get("/api/v1/federation", ListFederation)
	if os.Getenv("FEDERATION") == "true" {
		m.Post("/api/v1/federation/register", binding.Bind(EKCPServer{}), RegisterClusterToFederation)
		m.Delete("/api/v1/federation/:id", DeleteFederation)
		m.Get("/api/v1/federation/:id/info", InfoFederation)
	}
	if len(os.Getenv("FEDERATION_MASTER")) > 0 {
		SendRegistrationRequest()
	}

	// TODO: CRUD for routes
	// m.Post("/routes/new", binding.Bind(Route{}), NewRoute)
	// m.Delete("/routes/:id", DeleteRoute)
	// m.Get("/routes", ListRoutes)

	m.Run()
}

func GetProxyKubeConfig(ctx *macaron.Context) {
	id := ctx.Params(":id")
	var kubehost, host string

	// Check first if it was stored
	kubeconfig, err := Proxied.GetKubeConfig(id)
	if err == nil {
		ctx.PlainText(200, []byte(kubeconfig))
		return
	}

	if Federation.HasSlaves() {
		if cluster, err := Federation.Search(id); err == nil {
			ctx.PlainText(200, []byte(
				`apiVersion: v1
clusters:
- cluster:
    server: `+cluster.ProxyURL+`
  name: kind
contexts:
- context:
    cluster: kind
    user: kubernetes-admin
  name: kubernetes-admin@kind
current-context: kubernetes-admin@kind
kind: Config
preferences: {}`))
			return
		}

	}

	port, err := Proxied.GetProxy(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}

	kubehost = os.Getenv("KUBEHOST")
	host = os.Getenv("HOST")
	if len(kubehost) > 0 {
		host = kubehost
	}

	ctx.PlainText(200, []byte(
		`apiVersion: v1
clusters:
- cluster:
    server: http://`+host+":"+port+`
  name: kind
contexts:
- context:
    cluster: kind
    user: kubernetes-admin
  name: kubernetes-admin@kind
current-context: kubernetes-admin@kind
kind: Config
preferences: {}`))
}

func GetKubeEndpoint(ctx *macaron.Context) {
	id := ctx.Params(":id")

	port, err := Proxied.GetProxy(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}
	ctx.JSON(200, NewAPIResult(os.Getenv("HOST")+":"+port))
}

func KubePath(cluster string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "Could not find user home")
	}
	path := filepath.Join(usr.HomeDir, ".kube", "kind-config-"+cluster)

	return path, nil
}

func KubeConfig(id string) ([]byte, error) {
	// TODO: Check in others in fed mode
	res, err := KubePath(id)
	if err != nil {
		return []byte{}, err
	}

	dat, err := ioutil.ReadFile(res)
	if err != nil {

		return []byte{}, err
	}
	return dat, nil
}

func GetKubeConfig(ctx *macaron.Context) {
	id := ctx.Params(":id")

	res, err := KubeConfig(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error(), Output: string(res)})
		return
	}
	ctx.PlainText(200, res)
}

func ClusterInfo(ctx *macaron.Context) {
	id := ctx.Params(":id")

	if Federation.HasSlaves() {
		if cluster, err := Federation.Search(id); err == nil {
			ctx.JSON(200, cluster)
			return
		}
	}

	kubeC, err := GetClusterInfo(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}
	ctx.JSON(200, kubeC)
}

func InsertCluster(ctx *macaron.Context, kc KubernetesCluster) {
	config, err := kc.DecodeKubeConfig()
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}
	Proxied.AddKubeConfig(kc.Name, string(config))
	ctx.JSON(200, NewAPIResult("Cluster details stored"))
}

func NewCluster(ctx *macaron.Context, kc KubernetesCluster) {
	// TODO: In fed. mode - check availability and allocate by redirecting if necessary.

	if Federation.HasSlaves() {
		if err := Federation.Allocate(kc); err == nil {
			ctx.JSON(200, NewAPIResult("Cluster allocated correctly"))
			return
		}
	}

	if os.Getenv("FEDERATION") == "true" {
		ctx.JSON(500, APIResult{Error: "No available resources"})
		return
	}

	res, err := kc.Start()
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error(), Output: res})
		return
	}

	p, err := GetFreePort()
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}

	res, err = KubePath(kc.Name)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}

	err = KubeStartProxy(kc.Name, res, p)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}
	ctx.JSON(200, NewAPIResult(res))
}

func DeleteCluster(ctx *macaron.Context) {
	id := ctx.Params(":id")

	// Check first if it was stored
	_, err := Proxied.GetKubeConfig(id)
	if err == nil {
		Proxied.RemoveKubeConfig(id)
		ctx.JSON(200, NewAPIResult("Cluster deleted correctly"))
		return
	}

	// TODO: In fed. mode - check locally and propagate delete otherwise.
	if Federation.HasSlaves() {
		if err := Federation.Delete(id); err == nil {
			ctx.JSON(200, NewAPIResult("Cluster deleted correctly"))
			return
		}
	}

	res, err := Kind([]string{}, "delete", "cluster", "--name", id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error(), Output: res})
		return
	}

	err = Proxied.StopProxy(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error(), Output: res})
		return
	}

	ctx.JSON(200, NewAPIResult(res))
}

func ListClusters(ctx *macaron.Context) {
	ctx.JSON(200, NewAPIResult(""))
}

func ListFederation(ctx *macaron.Context) {
	ctx.JSON(200, Federation.Registered())
}

func DeleteFederation(ctx *macaron.Context) {
	id := ctx.ParamsInt(":id")

	if !Federation.HasSlaves() {

		ctx.JSON(200, Federation.Registered())
	}
	if err := Federation.Unregister(id); err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
	}

	ctx.JSON(200, Federation.Registered())
}

func InfoFederation(ctx *macaron.Context) {
	id := ctx.ParamsInt(":id")

	if !Federation.HasSlaves() {

		ctx.JSON(200, NewAPIResult("No endpoint registered"))
	}
	ep, err := Federation.Show(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
	}

	ctx.JSON(200, ep)
}
