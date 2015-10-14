package loadbalancers

import (
	"net"
	"time"

	"github.com/munnerz/gobalancer/pkg/api"
)

type Backend struct {
	Name         string
	IP           net.IP
	PollInterval time.Duration
	PollTimeout  time.Duration

	controlChan chan bool
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

func pollLoop(l LoadBalancer, b *Backend) {
	for {
		select {
		case _ = <-time.After(b.PollInterval):
			b.healthy = l.Poll(b)
		case _ = <-b.controlChan:
			return
		}
	}
}

func NewBackend(b *api.Backend) *Backend {
	return &Backend{
		Name:        b.Name,
		IP:          b.IP,
		controlChan: make(chan bool),
	}
}
