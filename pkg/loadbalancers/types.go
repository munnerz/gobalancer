package loadbalancers

import (
	"fmt"
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
)

var (
	types map[api.PortMapProtocol]func(name string, ip net.IP, portMap *api.PortMap, backends map[string]*Backend) LoadBalancer

	ErrBackendPanic        = fmt.Errorf("Panic in backend")
	ErrBackendsUnavailable = fmt.Errorf("No backends available")
	ErrLoadBalancerStopped = fmt.Errorf("LoadBalancer received stop signal")
)

func AddType(name api.PortMapProtocol, f func(string, net.IP, *api.PortMap, map[string]*Backend) LoadBalancer) {
	types[name] = f
}

func GetType(name api.PortMapProtocol) (func(string, net.IP, *api.PortMap, map[string]*Backend) LoadBalancer, error) {
	if t, ok := types[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("Unsupported loadbalancer protocol: %s", name)
}

func init() {
	types = make(map[api.PortMapProtocol]func(string, net.IP, *api.PortMap, map[string]*Backend) LoadBalancer)
}
