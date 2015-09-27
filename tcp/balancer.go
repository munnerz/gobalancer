package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
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
	Ports        []PortMap     `json:"ports"`
	Device       string        `json:"device"`
	Backends     Backends      `json:"backends"`
	PollInterval time.Duration `json:"poll_interval"`
	listeners    []net.Listener
	stopChan     chan bool
	stopReply    chan bool
	running      bool
}

type PortMap struct {
	Src int
	Dst int
}

func (p PortMap) String() string {
	return fmt.Sprintf("%d:%d", p.Src, p.Dst)
}

func (p *PortMap) UnmarshalJSON(d []byte) error {
	var j string

	err := json.Unmarshal(d, &j)

	if err != nil {
		return err
	}

	s := strings.Split(j, ":")

	if len(s) != 2 {
		return fmt.Errorf("Invalid portmap format (should be \"src:dst\")")
	}

	src, err1 := strconv.Atoi(s[0])
	dst, err2 := strconv.Atoi(s[1])

	if err1 != nil || err2 != nil {
		return fmt.Errorf("Invalid portmap format (should be \"src:dst\")")
	}

	p.Src, p.Dst = src, dst

	return nil
}

func (p *PortMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (t *LoadBalancer) Run(done chan *LoadBalancer) error {
	err := addressing.RegisterIP(t.IP, t.Subnet, t.Device)
	defer func() {
		err := addressing.UnregisterIP(t.IP, t.Subnet, t.Device)

		if err != nil {
			logging.Log(t.Name, log.Errorf, "Error unregistering IP address: %s", err.Error())
		}
	}()

	if err != nil && err != addressing.ErrAddressBound {
		logging.Log(t.Name, log.Errorf, "Error registering IP address: %s", err.Error())
		done <- t
		return err
	}

	go func() {
		for {
			for _, portmap := range t.Ports {
				for _, b := range t.Backends {
					b.poll(portmap.Dst)
				}
			}
			time.Sleep(t.PollInterval)
		}
	}()

	for _, portmap := range t.Ports {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", t.IP, portmap.Src))

		if err != nil {
			logging.Log(t.Name, log.Errorf, "Error binding portmap [%s]: %s", portmap, err.Error())
			done <- t
			return err
		}

		defer ln.Close()

		t.listeners = append(t.listeners, ln)

		go func(port int) {
			for {
				conn, err := ln.Accept()

				if err != nil {
					logging.Log(t.Name, log.Errorf, "Error accepting connection: %s", err.Error())
					break
				}

				logging.Log(t.Name, log.Debugf, "Accepted connection from %s", conn.RemoteAddr())

				go t.handleConnection(conn, port, 0)
			}
		}(portmap.Dst)
	}

	t.stopChan = make(chan bool)

	<-t.stopChan

	logging.Log(t.Name, log.Debugf, "Closing loadbalancer listener...")

	t.stopReply <- true
	done <- t

	return nil
}

func (t *LoadBalancer) Stop() {
	t.stopReply = make(chan bool)
	t.stopChan <- true
	<-t.stopReply
}

func (t *LoadBalancer) handleConnection(conn net.Conn, port int, retries int) {
	defer conn.Close()

	logging.Log(t.Name, log.Debugf, "Handling connection '%s'", conn.RemoteAddr())

	b, err := t.Backends.GetHealthyBackend()

	if err != nil {
		logging.Log(t.Name, log.Errorf, "Error getting backend: %s", err.Error())
		return
	}

	err = b.Proxy(conn, port)

	if err != nil {
		if err == ErrBackendFailed || err == ErrBackendPanic {
			logging.Log(t.Name, log.Errorf, "Error connecting to backend '%s:%d': %s", b.IP, port, err.Error())
			if retries < maxRetries {
				logging.Log(t.Name, log.Errorf, "%s retrying connection (%d)...", conn.RemoteAddr(), retries)
				t.handleConnection(conn, port, retries+1)
			}
			return
		}

	}

}
