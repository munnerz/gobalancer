package loadbalancers

import (
	"io"
	"net"

	"github.com/munnerz/gobalancer/pkg/api"
)

type LoadBalancer interface {
	Listen()
	NewConnection(*Backend) (net.Conn, error)
	Poll(*Backend) bool

	Backends() []*Backend
	ErrorChan() chan error
	ControlChan() chan bool
	ConnectionChan() chan net.Conn
}

func NewLoadBalancer(ip net.IP, port int, proto api.ServicePortType, backends []*api.Backend) (LoadBalancer, error) {
	f, err := GetType(proto)

	if err != nil {
		return nil, err
	}

	b := make([]*Backend, len(backends))

	for i, be := range backends {
		b[i] = &Backend{
			Backend: be,
		}
	}

	lb := f(ip, port, b)

	return lb, nil
}

func RunLoadBalancer(l LoadBalancer, errorChan chan error) {
	go l.Listen()

	for {
		select {
		case c := <-l.ConnectionChan():
			go proxy(l, c)
		case e := <-l.ErrorChan():
			errorChan <- e
			l.ControlChan() <- true
			return
		}
	}
}

func copyData(done chan error, src, dest net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			done <- ErrBackendPanic
		}
	}()

	_, err := io.Copy(src, dest)
	if err != nil {
		done <- err
		return
	}
	done <- nil
}

func proxy(l LoadBalancer, conn net.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrBackendPanic
		}
	}()

	backend, err := GetHealthyBackend(l)

	if err != nil {
		// No backends are available... give up with this connection
		conn.Close()
		return
	}

	bConn, err := l.NewConnection(backend)

	if err != nil {
		// Mark this backend as unhealthy and perform a healthcheck on it
		// in the background
		backend.healthy = false
		go healthCheck(l, backend)
		// Keep retrying until we're out of backends
		return proxy(l, conn)
	}

	defer bConn.Close()

	done1, done2 := make(chan error, 1), make(chan error, 1)

	go copyData(done1, conn, bConn)
	go copyData(done2, bConn, conn)

	select {
	case e := <-done1:
		err = e
	case e := <-done2:
		err = e
	}

	return
}
