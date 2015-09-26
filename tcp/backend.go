package tcp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Backend struct {
	IP          net.IP        `json:"ip"`
	Port        uint16        `json:"port"`
	Timeout     time.Duration `json:"timeout_duration"`
	healthy     bool
	connections map[*net.Conn]*net.Conn
	pollLock    sync.RWMutex
}

func (b *Backend) poll() {
	b.pollLock.Lock()
	defer b.pollLock.Unlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.IP, b.Port), b.Timeout)

	if err != nil {
		b.healthy = false
		return
	}

	defer conn.Close()

	b.healthy = true
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
