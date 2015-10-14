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
	Stop()

	Name() string
	Backends() []*Backend
	ErrorChan() chan error
	ControlChan() chan bool
	ConnectionChan() chan net.Conn
}

type LoadBalancerSpec struct {
	Name     string
	IP       net.IP
	PortMap  *api.PortMap
	Backends []*api.Backend
}

func NewLoadBalancer(l LoadBalancerSpec) (LoadBalancer, error) {
	f, err := GetType(l.PortMap.Protocol)

	if err != nil {
		return nil, err
	}

	b := make([]*Backend, len(l.Backends))

	for i, be := range l.Backends {
		b[i] = NewBackend(be)
	}

	lb := f(l.Name, l.IP, l.PortMap, b)

	return lb, nil
}

// TODO: There's currently no way to stop this function unless l returns an error.
// This'll mean that we cannot gracefully exit this loop through calling the
// loadbalancers Stop() method
func RunLoadBalancer(l LoadBalancer, errorChan chan error) {
	// Kick off periodically polling backends
	for _, b := range l.Backends() {
		go pollLoop(l, b)
	}

	go l.Listen()

	for {
		select {
		case c := <-l.ConnectionChan():
			go proxy(l, c)
		case err := <-l.ErrorChan():
			errorChan <- err
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
