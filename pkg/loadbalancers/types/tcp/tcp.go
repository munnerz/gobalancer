package tcp

import (
	"net"

	"github.com/munnerz/gobalancer/pkg/loadbalancers"
)

type TCP struct {
	ip   net.IP
	port int

	backends []*loadbalancers.Backend

	connectionChan chan net.Conn
	controlChan    chan bool
	errorChan      chan error
}

func NewTCP(ip net.IP, port int, backends []*loadbalancers.Backend) loadbalancers.LoadBalancer {
	return &TCP{
		ip:             ip,
		port:           port,
		backends:       backends,
		connectionChan: make(chan net.Conn),
		controlChan:    make(chan bool),
		errorChan:      make(chan error),
	}
}

func (t *TCP) ErrorChan() {
	return t.errorChan
}

func (t *TCP) ControlChan() {
	return t.controlChan
}

func (t *TCP) ConnectionChan() {
	return t.connectionChan
}

func init() {
	loadbalancers.AddType("tcp", NewTCP)
}
