package tcp

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/loadbalancers"
)

var (
	// ErrBackendPanic is a generic error for when a backend connection
	// panics whilst proxying
	ErrBackendPanic = fmt.Errorf("Panic in backend")
)

// Backend is a TCP backend instance
type backend struct {
	spec           api.Backend
	port           int
	healthy        bool
	connections    map[*net.Conn]*net.Conn
	connectionLock sync.Mutex
	pollLock       sync.Mutex
}

// IsHealthy returns true if this backend is able to accept connections,
// otherwise returns false
func (b *backend) IsHealthy() bool {
	return b.healthy
}

// Poll checks the health of the backend and updates the cached health status
func (b *backend) Poll() {
	b.pollLock.Lock()
	defer b.pollLock.Unlock()
	// TODO: Poll backend and update health
}

// Proxy proxies a connection to a backend. If this function panics, an error is
// returned instead of mitigating it
func (b *backend) Proxy(conn *net.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrBackendPanic
		}
	}()

	bc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.spec.IP, b.port))

	if err != nil {
		// TODO: Recheck health here?
		return err
	}

	defer bc.Close()

	done1, done2 := make(chan error, 1), make(chan error, 1)

	go proxy(done1, conn, &bc)
	go proxy(done2, &bc, conn)

	select {
	case e := <-done1:
		err = e
	case e := <-done2:
		err = e
	}

	return
}

func NewBackend(spec interface{}) (loadbalancers.Backend, error) {
	if s, ok := spec.(api.Backend); ok {
		return &backend{
			spec: s,
		}, nil
	}
	return nil, fmt.Errorf("Invalid Backend spec type")
}

func proxy(done chan error, src, dest *net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			done <- ErrBackendPanic
		}
	}()

	_, err := io.Copy(*src, *dest)
	if err != nil {
		done <- err
		return
	}
	done <- nil
}
