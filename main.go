package main

import (
	"fmt"
	"github.com/cssivision/reverseproxy"
	"github.com/go-macaron/binding"
	"go.uber.org/zap"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	kubeConfig "code.cloudfoundry.org/cf-operator/pkg/kube/config"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"k8s.io/kubectl/pkg/proxy"
)

type KindCluster struct {
	Name    string `form:"name" binding:"Required"`
	Version string `form:"version"` // TODO: Implement different kind cluster versions
}

type APIResult struct {
	Output string
	Error  string
}

type DB struct {
	Proxy map[string]string
}

var Proxied = &DB{Proxy: make(map[string]string)}

func main() {
	m := macaron.Classic()
	m.Use(macaron.Renderer())

	m.Get("/:id", GetKubeConfig)
	m.Post("/new", binding.Bind(KindCluster{}), NewCluster)
	m.Delete("/:id", DeleteCluster)
	m.Get("/", ListClusters)

	m.Get("/kube/:id", GetProxyPort)

	m.Run()
}
func GetFreePort() (int, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func GetProxyPort(ctx *macaron.Context) {
	id := ctx.Params(":id")
	port, ok := Proxied.Proxy[id]
	if !ok {
		ctx.JSON(500, APIResult{Error: "No such cluster has been proxied"})
		return
	}
	ctx.JSON(200, APIResult{Output: port})
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

	l, err := server.Listen("0.0.0.0", port)

	if err != nil {
		return err
	}
	fmt.Println("Starting to serve on %s\n", l.Addr().String())
	go server.ServeOnListener(l)
	Proxied.Proxy[clustername] = strconv.Itoa(port)

	return nil
}

func KubeConfigProxied(ctx *macaron.Context, w http.ResponseWriter, r *http.Request) {
	id := ctx.Params(":id")
	kubeconfig, _ := KubeConfig(id)

	reg, _ := regexp.Compile("localhost:(.*)")
	fmt.Println("Proxying " + id)
	reverse := reg.ReplaceAllString(string(kubeconfig), os.Getenv("HOST")+":"+os.Getenv("PORT")+"/kube/"+id)
	ctx.PlainText(200, []byte(reverse))
}

func KubeProxy(ctx *macaron.Context, w http.ResponseWriter, r *http.Request) {
	id := ctx.Params(":id")
	kubeconfig, _ := KubeConfig(id)

	reg, _ := regexp.Compile("localhost:(.*)")
	reverse := reg.FindString(string(kubeconfig))
	path, err := url.Parse("https://" + reverse)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Proxying " + id + " " + reverse)
	fmt.Println(path)
	reverseproxy.NewReverseProxy(path).ProxyHTTPS(w, r)

}

func KubePath(cluster string) (string, error) {
	res, err := Kind("get", "kubeconfig-path", "--name", cluster)
	if err != nil {

		return "", err
	}

	res = strings.TrimSuffix(res, "\n")
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
	if err != nil {
		return "", err
	}
	return string(out), nil
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
	ctx.JSON(200, APIResult{Output: res})
}

func DeleteCluster(ctx *macaron.Context) {
	id := ctx.Params(":id")
	res, err := Kind("delete", "cluster", "--name", id)
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error(), Output: res})

		return
	}
	ctx.JSON(200, APIResult{Output: res})
}

func ListClusters(ctx *macaron.Context) {
	res, err := Kind("get", "clusters")
	if err != nil {
		ctx.JSON(500, APIResult{Error: err.Error(), Output: res})

		return
	}
	ctx.JSON(200, APIResult{Output: res})
}
