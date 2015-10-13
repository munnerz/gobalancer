package loadbalancers

import (
	"fmt"
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
)

var (
	types map[api.ServicePortType]func(ip net.IP, port int, backends []*Backend) LoadBalancer

	ErrBackendPanic        = fmt.Errorf("Panic in backend")
	ErrBackendsUnavailable = fmt.Errorf("No backends available")
)

func AddType(name api.ServicePortType, f func(net.IP, int, []*Backend) LoadBalancer) {
	types[name] = f
}

func GetType(name api.ServicePortType) (func(net.IP, int, []*Backend) LoadBalancer, error) {
	if t, ok := types[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("Unsupported loadbalancer protocol: %s", name)
}

func init() {
	types = make(map[api.ServicePortType]func(net.IP, int, []*Backend) LoadBalancer)
}
