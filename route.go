package main

import (
	"fmt"
	nats "github.com/nats-io/nats.go"
	macaron "gopkg.in/macaron.v1"
)

var DefaultRouteRegister *RouteRegister

type Route struct {
	Host    string `form:"host" binding:"Required"`
	Domain  string `form:"domain"`
	Port    string `form:"port"`
	TLSPort string `form:"tls_port"`
}

type RouteRegister struct {
	Routes []Route
	Nats   *nats.Conn
}

func NewRouteRegister() (*RouteRegister, error) {
	if DefaultRouteRegister == nil {
		// Connect to a server
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			return nil, err
		}
		DefaultRouteRegister = &RouteRegister{Nats: nc}
	}

	return DefaultRouteRegister, nil
}

func (rr *RouteRegister) Register(r Route) error {

	var err error
	d := fmt.Sprintf("%+q", []string{r.Domain, "*." + r.Domain}) // need to support multiple?
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
	for _, cluster := range result.Clusters {
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
	err = rr.Register(Route{Host: ip, Port: "80", Domain: route})
	if err != nil {
		return err
	}

	err = rr.Register(Route{Host: ip, TLSPort: "443", Domain: route})
	if err != nil {
		return err
	}

	return nil
}
