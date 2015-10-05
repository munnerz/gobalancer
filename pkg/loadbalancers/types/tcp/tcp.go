package tcp

import (
	"fmt"
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/loadbalancers"
)

const (
	name = "tcp"
)

// LoadBalancer contains the runtime state of a TCP loadbalancer
type LoadBalancer struct {
	ip          net.IP
	port        api.ServicePort
	backends    []backend
	controlChan chan bool
}

// Run will launch the TCP loadbalancer and begin accepting and proxying connections
func (l *LoadBalancer) Run() error {

	connChan := make(chan net.Conn)
	listenerControlChan := make(chan bool)
	listenerErrChan := make(chan error)

	go l.listen(connChan, listenerControlChan, listenerErrChan)

Loop:
	for {
		select {
		case conn := <-connChan:
			go l.accept(&conn)
		case <-l.controlChan:
			break Loop
		case <-listenerErrChan:
			break Loop
		}
	}

	listenerControlChan <- true

	return nil
}

// Stop sends a signal to the Run function to stop accepting connections
func (l *LoadBalancer) Stop() {
	l.controlChan <- true
}

func NewLoadBalancer(spec interface{}) (loadbalancers.LoadBalancer, error) {
	if s, ok := spec.(api.LoadBalancer); ok {
		backends := make([]*backend, len(s.Backends))

		for i, b := range s.Backends {
			be, err := NewBackend(b)

			if err != nil {
				return nil, err
			}

			backends[i] = be.(*backend)
		}

		return &LoadBalancer{
			spec:        s,
			controlChan: make(chan bool),
		}, nil
	}
	return nil, fmt.Errorf("Invalid TCPLoadBalancer spec type")
}

func init() {
	loadbalancers.AddType(name, NewLoadBalancer)
}
