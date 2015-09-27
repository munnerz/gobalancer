package tcp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrBackendsUnavailable = errors.New("No backends available")
	ErrBackendFailed       = errors.New("Unexpected backend failure")
	ErrBackendPanic        = errors.New("Panic in backend")
)

type Backends []*Backend

type Backend struct {
	Name           string        `json:"name"`
	IP             net.IP        `json:"ip"`
	Timeout        time.Duration `json:"poll_timeout"`
	healthy        bool
	connections    map[*net.Conn]*net.Conn
	connectionLock sync.RWMutex
	pollLock       sync.RWMutex
}

func proxy(done chan error, src, dest net.Conn) {
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

func (b *Backend) Proxy(conn net.Conn, port int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrBackendPanic
		}
	}()

	bc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.IP, port))

	if err != nil {
		b.pollLock.Lock()
		defer b.pollLock.Unlock()
		b.healthy = false
		return ErrBackendFailed
	}

	defer bc.Close()

	if conn == nil || bc == nil {
		panic("Connections have gone away!")
	}

	b.addConnection(&conn, &bc)
	defer b.deleteConnection(&conn)

	done1, done2 := make(chan error, 1), make(chan error, 1)

	go proxy(done1, conn, bc)
	go proxy(done2, bc, conn)

	select {
	case e := <-done1:
		err = e
	case e := <-done2:
		err = e
	}

	return
}

func (l *Backends) GetHealthyBackend() (*Backend, error) {
	b := l.leastconn()

	if b == nil {
		return nil, ErrBackendsUnavailable
	}

	return b, nil
}

func (l *Backends) leastconn() *Backend {
	// Implements a least connection based load balancing algorithm
	var b *Backend

	for _, be := range *l {
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

	return b
}

func (b *Backend) poll(port int) bool {
	b.pollLock.Lock()
	defer b.pollLock.Unlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.IP, port), b.Timeout)

	if err != nil {
		b.healthy = false
		return false
	}

	defer conn.Close()

	b.healthy = true
	return true
}

func (b *Backend) addConnection(source, target *net.Conn) {
	b.connectionLock.Lock()
	defer b.connectionLock.Unlock()
	if b.connections == nil {
		b.connections = make(map[*net.Conn]*net.Conn)
	}
	b.connections[source] = target
}

func (b *Backend) deleteConnection(source *net.Conn) {
	b.connectionLock.Lock()
	defer b.connectionLock.Unlock()
	delete(b.connections, source)
}
