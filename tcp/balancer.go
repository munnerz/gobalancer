package tcp

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/gobalancer/logging"
)

const (
	maxRetries = 10
)

type LoadBalancer struct {
	Name         string        `json:"name"`
	IP           net.IP        `json:"ip"`
	Port         uint16        `json:"port"`
	Backends     Backends      `json:"backends"`
	PollInterval time.Duration `json:"poll_interval"`
}

func (t *LoadBalancer) Run(done chan error) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))

	if err != nil {
		done <- err
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

	b, err := t.Backends.GetHealthyBackend()

	if err != nil {
		logging.Log(t.Name, log.Errorf, "Error getting backend: %s", err.Error())
		return
	}

	if s, err := b.Proxy(conn); !s {
		logging.Log(t.Name, log.Errorf, "Error connecting to backend '%s:%d': %s", b.IP, b.Port, err.Error())
		if retries < maxRetries {
			logging.Log(t.Name, log.Errorf, "%s retrying connection (%d)...", conn.RemoteAddr(), retries)
			t.handleConnection(conn, retries+1)
		}
		return
	}
}
