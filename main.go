package main

import (
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"os"
)

type KindCluster struct {
	Name    string `form:"name" binding:"Required"`
	Version string `form:"version"` // TODO: Implement different kind cluster versions
}

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
	m.Post("/new", binding.Bind(KindCluster{}), NewCluster)
	m.Delete("/:id", DeleteCluster)
	m.Get("/", ListClusters)

	m.Get("/kubeconfig/:id", GetProxyKubeConfig)
	m.Get("/kube/:id", GetKubeEndpoint)

	// TODO: CRUD for routes
	// m.Post("/routes/new", binding.Bind(Route{}), NewRoute)
	// m.Delete("/routes/:id", DeleteRoute)
	// m.Get("/routes", ListRoutes)

	m.Run()
}

func GetProxyKubeConfig(ctx *macaron.Context) {
	id := ctx.Params(":id")
	port, err := Proxied.GetProxy(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}

	kubehost := os.Getenv("KUBEHOST")
	host := os.Getenv("HOST")
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
	res, err := Kind("get", "kubeconfig-path", "--name", cluster)
	if err != nil {

		return "", err
	}

	return res, nil
}
func KubeConfig(id string) ([]byte, error) {
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

func NewCluster(ctx *macaron.Context, kc KindCluster) {
	res, err := Kind("create", "cluster", "--name", kc.Name)
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
	res, err := Kind("delete", "cluster", "--name", id)
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
