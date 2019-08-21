package main

import (
	"fmt"
	nats "github.com/nats-io/nats.go"
	macaron "gopkg.in/macaron.v1"
	"os"
)

type RouteRegister struct{}

func NewRouteRegister() *RouteRegister {
	return &RouteRegister{}
}

func (rr *RouteRegister) Register(host, port string, domains []string) error {
	// Connect to a server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}

	d := fmt.Sprintf("%+q", domains)
	err = nc.Publish("gorouter.register", []byte(`{"host":"`+host+`","port":`+port+`, "uris": `+d+` }`))
	if err != nil {
		return err
	}
	return nil
}

func MacaronRR() macaron.Handler {
	return func(ctx *macaron.Context) {
		result := NewAPIResult("")
		for _, cluster := range result.Clusters {
			RegisterCluster(cluster)
		}

	}
}

func RegisterCluster(clustername string) {
	listenIP := os.Getenv("HOST")
	rr := NewRouteRegister()

	ip, _ := GetKubeIP(clustername)

	rr.Register(ip, "80", []string{clustername + "." + listenIP + ".nip.io"})
	rr.Register(ip, "443", []string{clustername + "." + listenIP + ".nip.io"})
}
