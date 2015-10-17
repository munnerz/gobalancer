package tcp

import (
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/loadbalancers"
)

type TCP struct {
	name    string
	ip      net.IP
	portMap *api.PortMap

	backends map[string]*loadbalancers.Backend

	connectionChan chan net.Conn
	controlChan    chan bool
	errorChan      chan error
}

func NewTCP(name string, ip net.IP, portMap *api.PortMap, backends map[string]*loadbalancers.Backend) loadbalancers.LoadBalancer {
	return &TCP{
		name:           name,
		ip:             ip,
		portMap:        portMap,
		backends:       backends,
		connectionChan: make(chan net.Conn),
		controlChan:    make(chan bool),
		errorChan:      make(chan error),
	}
}

func (t *TCP) Stop() {
	t.controlChan <- true
}

func (t *TCP) Name() string {
	return t.name
}

func (t *TCP) Backends() map[string]*loadbalancers.Backend {
	return t.backends
}

func (t *TCP) ErrorChan() chan error {
	return t.errorChan
}

func (t *TCP) ControlChan() chan bool {
	return t.controlChan
}

func (t *TCP) ConnectionChan() chan net.Conn {
	return t.connectionChan
}

func init() {
	loadbalancers.AddType("tcp", NewTCP)
}
