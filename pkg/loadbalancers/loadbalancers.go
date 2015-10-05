package loadbalancers

import (
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
)

var (
	loadbalancers map[string](func(interface{}) (LoadBalancer, error))
)

type LB struct {
	ServicePort *api.ServicePort
	IP          *net.IPNet
	Device      *string
	Backends    []*api.Backend
}

// LoadBalancer is a generic interface for LoadBalancers
type LoadBalancer interface {
	Run() error
	Stop()
}

// Backend is a generic backend type for loadbalancers
type Backend interface {
	IsHealthy() bool
}

// AddType registers a type of loadbalancer with the application
func AddType(name string, l func(interface{}) (LoadBalancer, error)) {
	loadbalancers[name] = l
}

func init() {
	loadbalancers = make(map[string]func(interface{}) (LoadBalancer, error))
}
