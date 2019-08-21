package main
import (
	nats "github.com/nats-io/nats.go"
"fmt"
)

type RouteRegister struct{}

func NewRouteRegister() *RouteRegister {
	return &RouteRegister{}
}


func (rr *RouteRegister) Register(host,port string, domains []string) error {
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