package tcp

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	maxRetries = 10
)

type LoadBalancer struct {
	Name         string        `json:"name"`
	IP           net.IP        `json:"ip"`
	Port         uint16        `json:"port"`
	Backends     []*Backend    `json:"backends"`
	PollInterval time.Duration `json:"poll_interval"`
}

func (t *LoadBalancer) Run() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))

	if err != nil {
		return err
	}

	go func() {
		for {
			for _, b := range t.Backends {
				b.poll()
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

		go t.handleConnection(conn, 0)
	}
}

func (t *LoadBalancer) Stop() error {
	return nil
}

func (t *LoadBalancer) handleConnection(conn net.Conn, retries int) {
	defer conn.Close()

	b, err := t.getActiveBackend()

	if err != nil {
		t.log(log.Errorf, "Error handling connection: %s", err.Error())
		return
	}

	bc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.IP, b.Port))

	if err != nil {
		go b.poll()
		t.log(log.Errorf, "Error connecting to backend '%s:%d': %s", b.IP, b.Port, err.Error())
		if retries < maxRetries {
			t.log(log.Errorf, "%s retrying connection (%d)...", conn.RemoteAddr(), retries)
			t.handleConnection(conn, retries+1)
		}
		return
	}

	defer bc.Close()

	t.log(log.Printf, "%s->%s: opening...", conn.RemoteAddr(), bc.RemoteAddr())

	b.addConnection(&conn, &bc)

	done := make(chan error)

	go proxy(done, bc.Read, conn.Write)
	go proxy(done, conn.Read, bc.Write)

	err = <-done

	b.deleteConnection(&conn)

	if err != nil {
		t.log(log.Errorf, "%s->%s: %s", conn.RemoteAddr(), bc.RemoteAddr(), err.Error())
		return
	}

}

func (t *LoadBalancer) log(f func(string, ...interface{}), s string, sf ...interface{}) {
	f("%s: %v", t.Name, fmt.Sprintf(s, sf...))
}

func (t *LoadBalancer) getActiveBackend() (*Backend, error) {
	// Implements a least connection based load balancing algorithm
	var b *Backend

	for _, be := range t.Backends {
		go log.Debugf("Backend '%s:%d' active connections: %d", be.IP, be.Port, len(be.connections))

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
		return nil, fmt.Errorf("No backend available!")
	}

	return b, nil
}

func proxy(done chan error, r func([]byte) (int, error), w func([]byte) (int, error)) {
	for {
		data := make([]byte, 256)
		nr, err := r(data)

		if err != nil {
			w(data[0:nr])
			done <- err
			return
		}

		w(data[0:nr])
	}
}
