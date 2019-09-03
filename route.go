package main

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	macaron "gopkg.in/macaron.v1"
)

var DefaultRouteRegister *RouteRegister

type Route struct {
	Host    string `form:"host" binding:"Required"`
	Domain  string `form:"domain"`
	Port    string `form:"port"`
	TLSPort string `form:"tls_port"`
	Cluster string
}

func (r Route) ToString() string {
	return fmt.Sprintf("%s-%s-%s-%s", r.Host, r.Domain, r.Port, r.TLSPort)
}

type RouteRegister struct {
	Routes map[string]map[string]Route
	Nats   *nats.Conn
}

func NewRouteRegister() (*RouteRegister, error) {
	if DefaultRouteRegister == nil {
		// Connect to a server
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			return nil, err
		}
		DefaultRouteRegister = &RouteRegister{Nats: nc, Routes: make(map[string]map[string]Route)}
	}

	return DefaultRouteRegister, nil
}

func (rr *RouteRegister) ClusterRoutes(clustername string) ([]Route, error) {

	routes, ok := rr.Routes[clustername]
	if !ok {
		return []Route{}, errors.New("No routes found for clustername")
	}

	var res []Route
	for _, r := range routes {
		res = append(res, r)
	}
	return res, nil
}
func (rr *RouteRegister) Register(r Route) error {
	if _, ok := rr.Routes[r.Cluster]; !ok {
		rr.Routes[r.Cluster] = make(map[string]Route)
	}
	rr.Routes[r.Cluster][r.ToString()] = r

	var err error
	d := fmt.Sprintf("%+q", []string{r.Domain}) // need to support multiple?
	if len(r.TLSPort) > 0 {
		err = rr.Nats.Publish("router.register", []byte(`{"host":"`+r.Host+`", "tls_port": `+r.TLSPort+`, "uris": `+d+`, "tags":{"type":"cluster"} }`))
	} else {
		err = rr.Nats.Publish("router.register", []byte(`{"host":"`+r.Host+`", "port":`+r.Port+`, "uris": `+d+`, "tags":{"type":"cluster"} }`))
	}
	if err != nil {
		return err
	}
	return nil
}

func MacaronRR(domain string) macaron.Handler {
	return func(ctx *macaron.Context) {
		RegisterAll(domain)
	}
}

func RegisterAll(domain string) {
	result := NewAPIResult("")
	for cluster, _ := range result.ActiveEndpoints {
		err := RegisterCluster(cluster, domain)
		if err != nil {
			fmt.Println("[WARN] Failed registering route for", cluster, err.Error())
		}
	}
}

func RegisterCluster(clustername, domain string) error {
	rr, err := NewRouteRegister()
	if err != nil {
		return err
	}
	ip, err := GetKubeIP(clustername)
	if err != nil {
		return err
	}
	// TODO: Support other ports? (and maybe TCP conns as well?)
	//route:= clustername + "." + listenIP + ".nip.io"
	route := clustername + "." + domain

	fmt.Println("[INFO] Registering route", route)
	err = rr.Register(Route{Host: ip, Port: "80", Domain: "*." + route, Cluster: clustername})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, TLSPort: "443", Domain: "*." + route, Cluster: clustername})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, Port: "80", Domain: route, Cluster: clustername})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, TLSPort: "443", Domain: route, Cluster: clustername})
	if err != nil {
		return err
	}

	route = clustername + "." + ip + "." + domain

	fmt.Println("[INFO] Registering route", route)
	err = rr.Register(Route{Host: ip, Port: "80", Domain: "*." + route, Cluster: clustername})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, TLSPort: "443", Domain: "*." + route, Cluster: clustername})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, Port: "80", Domain: route, Cluster: clustername})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, TLSPort: "443", Domain: route, Cluster: clustername})
	if err != nil {
		return err
	}

	return nil
}
