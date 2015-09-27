package tcp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Backends []*Backend

type Backend struct {
	Name        string        `json:"name"`
	IP          net.IP        `json:"ip"`
	Port        uint16        `json:"port"`
	Timeout     time.Duration `json:"poll_timeout"`
	healthy     bool
	connections map[*net.Conn]*net.Conn
	pollLock    sync.RWMutex
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

func (b *Backend) Proxy(conn net.Conn) error {
	bc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.IP, b.Port))

	if err != nil {
		b.pollLock.Lock()
		defer b.pollLock.Unlock()
		b.healthy = false
		return err
	}

	defer bc.Close()

	b.addConnection(&conn, &bc)

	done := make(chan error)

	go proxy(done, bc.Read, conn.Write)
	go proxy(done, conn.Read, bc.Write)

	<-done

	b.deleteConnection(&conn)

	return nil
}

func (l *Backends) GetHealthyBackend() (*Backend, error) {
	b := l.leastconn()

	if b == nil {
		l.quickPoll()
		b = l.leastconn()
		if b == nil {
			return nil, fmt.Errorf("No backend available")
		}
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

func (l *Backends) quickPoll() {
	for _, b := range *l {
		if b.poll() {
			return
		}
	}
}

func (b *Backend) poll() bool {
	b.pollLock.Lock()
	defer b.pollLock.Unlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.IP, b.Port), b.Timeout)

	if err != nil {
		b.healthy = false
		return false
	}

	defer conn.Close()

	b.healthy = true
	return true
}

func (b *Backend) addConnection(source, target *net.Conn) {
	if b.connections == nil {
		b.connections = make(map[*net.Conn]*net.Conn)
	}
	b.connections[source] = target
}

func (b *Backend) deleteConnection(source *net.Conn) {
	delete(b.connections, source)
}
