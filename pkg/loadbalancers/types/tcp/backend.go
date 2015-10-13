package tcp

import (
	"fmt"
	"net"

	"github.com/munnerz/gobalancer/pkg/loadbalancers"
)

// Poll checks the health of the backend and updates the cached health status
func (t *TCP) Poll(b *loadbalancers.Backend) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.IP, t.port), b.PollTimeout)

	if err != nil {
		return false
	}

	defer conn.Close()

	return true
}

func (t *TCP) NewConnection(b *loadbalancers.Backend) (net.Conn, error) {
	return net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.IP, t.port), b.PollTimeout)
}
