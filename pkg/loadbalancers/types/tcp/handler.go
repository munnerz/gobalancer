package tcp

import (
	"fmt"
)

var (
	ErrBackendsUnavailable = fmt.Errorf("No backends available")
)

func (l *LoadBalancer) getHealthyBackend() (*backend, error) {
	// Implements a least connection based load balancing algorithm
	var b *backend

	for _, be := range l.backends {
		// Skip unhealthy backends
		if !be.healthy {
			continue
		}

		if b == nil {
			b = &be
			continue
		}

		// Choose the backend with the least connections
		if len(be.connections) < len(b.connections) {
			b = &be
		}
	}

	if b == nil {
		return nil, ErrBackendsUnavailable
	}

	return b, nil
}

func (l *LoadBalancer) accept(conn *connection) {

	// Get a healthy backend
	// Call backend.Proxy(conn)
}
