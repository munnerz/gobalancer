package loadbalancers

import (
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
)

type Backend struct {
	*api.Backend

	healthy     bool
	connections []*net.Conn
}

func healthCheck(l LoadBalancer, b *Backend) {
	b.healthy = l.Poll(b)
}

func GetHealthyBackend(l LoadBalancer) (*Backend, error) {
	// Implements a least connection based load balancing algorithm
	var b *Backend

	for _, be := range l.Backends() {
		// Skip unhealthy backends
		if !be.healthy {
			continue
		}

		if b == nil {
			b = be
			continue
		}

		// Choose the backend with the least connections
		if len(be.connections) < len(b.connections) {
			b = be
		}
	}

	if b == nil {
		return nil, ErrBackendsUnavailable
	}

	return b, nil
}
