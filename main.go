package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
	"time"
)

type LoadBalancer interface {
	Run() error
	Stop() error
}

type TCPLoadBalancer struct {
	Name         string        `json:"name"`
	IP           net.IP        `json:"ip"`
	Port         uint16        `json:"port"`
	Backends     []*Backend    `json:"backends"`
	PollInterval time.Duration `json:"poll_interval"`
}

type Backend struct {
	IP          net.IP        `json:"ip"`
	Port        uint16        `json:"port"`
	Timeout     time.Duration `json:"timeout_duration"`
	health      bool
	connections map[*net.Conn]*net.Conn
}

func (t *TCPLoadBalancer) getActiveBackend() (*Backend, error) {
	// Implements a least connection based load balancing algorithm
	var b *Backend

	for _, be := range t.Backends {
		log.Debugf("Backend '%s:%d' active connections: %d", be.IP, be.Port, len(be.connections))

		// Skip unhealthy backends
		if !be.health {
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
		return nil, fmt.Errorf("No backend available!")
	}

	return b, nil
}

func (t *TCPLoadBalancer) errorf(s string, sf ...interface{}) {
	log.Errorf("%s: %v", t.Name, fmt.Sprintf(s, sf...))
}

func proxy(done chan error, r func([]byte) (int, error), w func([]byte) (int, error)) {
	for {
		data := make([]byte, 256)
		nr, err := r(data)

		if err != nil {
			w(data[0:nr])
			done <- err
		}

		w(data[0:nr])
	}
}

func poll(ip net.IP, port uint16, timeout time.Duration) (bool, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)

	if err != nil {
		return false, err
	}

	defer conn.Close()

	return true, nil
}

func (t *TCPLoadBalancer) handleConnection(conn net.Conn) {
	defer conn.Close()

	b, err := t.getActiveBackend()

	if err != nil {
		t.errorf("Error handling connection: %s", err.Error())
		return
	}

	bc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.IP, b.Port))

	if err != nil {
		t.errorf("Error connecting to backend '%s:%d': %s", b.IP, b.Port, err.Error())
		return
	}

	defer bc.Close()

	b.connections[&conn] = &bc

	done := make(chan error)

	go proxy(done, bc.Read, conn.Write)
	go proxy(done, conn.Read, bc.Write)

	err = <-done

	delete(b.connections, &conn)

	if err != nil {
		t.errorf("Error proxying connection to backend '%s:%d': %s", b.IP, b.Port, err.Error())
		return
	}

	return
}

func (t *TCPLoadBalancer) Run() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))

	if err != nil {
		return err
	}

	go func() {
		for {
			for _, b := range t.Backends {
				s, _ := poll(b.IP, b.Port, b.Timeout)

				b.health = s

				log.Debugf("Backend '%s:%d' status: %v", b.IP, b.Port, s)

				if err != nil {
					t.errorf("Error polling backend '%s:%d': %s", b.IP, b.Port, err.Error())
					continue
				}
			}
			time.Sleep(t.PollInterval)
		}
	}()

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Errorf("Error accepting connection: %s", err.Error())
			continue
		}

		go t.handleConnection(conn)
	}
}

func (t *TCPLoadBalancer) Stop() error {
	return nil
}

func main() {
	t := TCPLoadBalancer{
		Name: "Test",
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 2000,
		Backends: []*Backend{
			&Backend{
				IP:          net.IPv4(127, 0, 0, 1),
				Port:        10000,
				connections: make(map[*net.Conn]*net.Conn),
			},
			&Backend{
				IP:          net.IPv4(127, 0, 0, 1),
				Port:        10001,
				connections: make(map[*net.Conn]*net.Conn),
			},
			&Backend{
				IP:          net.IPv4(127, 0, 0, 1),
				Port:        10002,
				connections: make(map[*net.Conn]*net.Conn),
			},
		},
		PollInterval: time.Millisecond * 1000,
	}

	log.SetLevel(log.DebugLevel)

	t.Run()
}
