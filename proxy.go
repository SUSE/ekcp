package main

import (
	"fmt"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/kubectl/pkg/proxy"
	"os"
	"strconv"
	"time"

	kubeConfig "code.cloudfoundry.org/cf-operator/pkg/kube/config"
)

var Proxied = &DB{Endpoints: make(map[string]*Proxy)}

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

func GetFreePort() (int, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func KubeStartProxy(clustername, kubeconfig string, port int) error {
	listenIP := os.Getenv("HOST")

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

	l, err := server.Listen(listenIP, port)

	if err != nil {
		return err
	}

	retval, err := NewStoppableListener(l)

	if err != nil {
		return err
	}

	fmt.Println("Starting to serve on", l.Addr().String())
	go server.ServeOnListener(retval)
	Proxied.SetProxy(clustername, strconv.Itoa(port), server, retval)

	return nil
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
