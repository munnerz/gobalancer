package tcp

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/gobalancer/addressing"
	"github.com/munnerz/gobalancer/logging"
)

const (
	maxRetries = 10
)

type LoadBalancer struct {
	Name         string        `json:"name"`
	IP           net.IP        `json:"ip"`
	Subnet       net.IP        `json:"subnet"`
	Port         uint16        `json:"port"`
	Device       string        `json:"device"`
	Backends     Backends      `json:"backends"`
	PollInterval time.Duration `json:"poll_interval"`
	mtu          int
}

func (t *LoadBalancer) Run(done chan error) error {
	err := addressing.RegisterIP(t.IP, t.Subnet, t.Device)

	if err != nil && err != addressing.ErrAddressBound {
		done <- err
		return err
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", t.IP, t.Port))

	if err != nil {
		done <- err
		return err
	}

	defer ln.Close()

	go func() {
		for {
			for _, b := range t.Backends {
				b.poll()
			}
			time.Sleep(t.PollInterval)
		}
	}()

	logging.Log(t.Name, log.Debugf, "Entering connection loop...")

	for {
		conn, err := ln.Accept()

		if err != nil {
			logging.Log(t.Name, log.Errorf, "Error accepting connection: %s", err.Error())
			done <- err
			break
		}

		logging.Log(t.Name, log.Debugf, "Accepted connection from %s", conn.RemoteAddr())

		go t.handleConnection(conn, 0)
	}

	logging.Log(t.Name, log.Debugf, "Closing loadbalancer listener...")

	done <- nil
	return nil
}

func (t *LoadBalancer) Stop() error {
	err := addressing.UnregisterIP(t.IP, t.Subnet, t.Device)

	if err != nil {
		return err
	}

	return nil
}

func (t *LoadBalancer) handleConnection(conn net.Conn, retries int) {
	defer conn.Close()

	logging.Log(t.Name, log.Debugf, "Handling connection '%s'", conn.RemoteAddr())

	b, err := t.Backends.GetHealthyBackend()

	if err != nil {
		logging.Log(t.Name, log.Errorf, "Error getting backend: %s", err.Error())
		return
	}

	err = b.Proxy(conn)

	if err != nil {
		if err == ErrBackendFailed || err == ErrBackendPanic {
			logging.Log(t.Name, log.Errorf, "Error connecting to backend '%s:%d': %s", b.IP, b.Port, err.Error())
			if retries < maxRetries {
				logging.Log(t.Name, log.Errorf, "%s retrying connection (%d)...", conn.RemoteAddr(), retries)
				t.handleConnection(conn, retries+1)
			}
			return
		}

	}

}
