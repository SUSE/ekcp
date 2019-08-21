package main

import (
	"fmt"
	"github.com/go-macaron/binding"
	"go.uber.org/zap"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	kubeConfig "code.cloudfoundry.org/cf-operator/pkg/kube/config"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"k8s.io/kubectl/pkg/proxy"
)

type KindCluster struct {
	Name     string `form:"name" binding:"Required"`
	Version  string `form:"version"` // TODO: Implement different kind cluster versions
	Register bool   `form:"route-register"`
	Route    string `form:"route"`
}

type APIResult struct {
	Output          string
	Clusters        []string
	ActiveEndpoints map[string]string
	Error           string
}

type Proxy struct {
	Port     string
	Server   *proxy.Server
	Listener *StoppableListener
}

type DB struct {
	Endpoints map[string]*Proxy
}

func (d *DB) GetProxy(s string) (string, error) {
	p, ok := d.Endpoints[s]
	if !ok {
		return "", errors.New("No Proxy found for " + s)
	}

	return p.Port, nil

}

func (d *DB) SetProxy(id, port string, server *proxy.Server, listener *StoppableListener) {
	d.Endpoints[id] = &Proxy{Port: port, Server: server, Listener: listener}
}

func (d *DB) StopProxy(id string) error {
	p, ok := d.Endpoints[id]
	if !ok {
		return errors.New("No Proxy found for " + id)
	}
	p.Listener.Stop()
	delete(d.Endpoints, id)
	return nil
}

var Proxied = &DB{Endpoints: make(map[string]*Proxy)}

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
	// Get active endpoints
	for cluster, p := range Proxied.Endpoints {
		activeEndpoints[cluster] = os.Getenv("HOST") + ":" + p.Port
	}

	return APIResult{Clusters: clusters, ActiveEndpoints: activeEndpoints, Output: output}
}

func ProxyStartup() error {
	result := NewAPIResult("")
	for _, cluster := range result.Clusters {
		kubeconfigPath, err := KubePath(cluster)
		if err != nil {
			return err
		}
		port, err := GetFreePort()
		if err != nil {
			return err
		}
		err = KubeStartProxy(cluster, kubeconfigPath, port)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	m := macaron.Classic()
	err := ProxyStartup()
	if err != nil {
		panic(err)
	}
	m.Use(macaron.Renderer())

	m.Get("/:id", GetKubeConfig)
	m.Post("/new", binding.Bind(KindCluster{}), NewCluster)
	m.Delete("/:id", DeleteCluster)
	m.Get("/", ListClusters)

	m.Get("/kubeconfig/:id", GetProxyKubeConfig)
	m.Get("/kube/:id", GetKubeEndpoint)

	m.Run()
}
func GetFreePort() (int, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func GetProxyKubeConfig(ctx *macaron.Context) {
	id := ctx.Params(":id")
	port, err := Proxied.GetProxy(id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error()})
		return
	}

	ctx.PlainText(200, []byte(
`apiVersion: v1
clusters:
- cluster:
    server: http://`+os.Getenv("HOST")+":"+port+`
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

func KubeStartProxy(clustername, kubeconfig string, port int) error {
	z, e := zap.NewProduction()
	if e != nil {
		return errors.New("Cannot create logger")
	}
	defer z.Sync() // flushes buffer, if any
	sugar := z.Sugar()

	restConfig, err := kubeConfig.NewGetter(sugar).Get(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "Could not connect with kubeconfig")
	}

	server, err := proxy.NewServer("", "/", "/", nil, restConfig, 90*time.Second)

	// Separate listening from serving so we can report the bound port
	// when it is chosen by os (eg: port == 0)

	l, err := server.Listen(os.Getenv("HOST"), port)

	if err != nil {
		return err
	}

	retval, err := NewStoppableListener(l)

	if err != nil {
		return err
	}

	fmt.Println("Starting to serve on %s\n", l.Addr().String())
	go server.ServeOnListener(retval)
	Proxied.SetProxy(clustername, strconv.Itoa(port), server, retval)

	return nil
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

func Kind(args ...string) (string, error) {
	out, err := exec.Command("kind", args...).CombinedOutput()
	output := string(out)
	output = strings.TrimSuffix(output, "\n")

	if err != nil {
		return output, err
	}
	return output, nil
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
