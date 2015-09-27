package tcp

import (
	"net"
	"runtime"
	"testing"
	"time"
)

const (
	initialText  = "Initial"
	responseText = "Response"
)

var (
	initialTextBytes  = []byte(initialText)
	responseTextBytes = []byte(responseText)
)

func runBackendServer(d chan error, t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:32145")

	if err != nil {
		t.Errorf("Error binding socket: %s", err.Error())
	}

	for {
		conn, err := ln.Accept()
		defer conn.Close()

		go func() {
			if err != nil {
				//t.Errorf("Error accepting connection: %s", err.Error())
				return
			}

			n, err := conn.Write(initialTextBytes)

			if err != nil {
				//t.Errorf("Error writing to test client: %s", err.Error())
				return
			}

			if n != len(initialTextBytes) {
				//t.Errorf("Could not write complete test message")
				return
			}

			b := make([]byte, len(responseTextBytes))
			n, err = conn.Read(b)

			if err != nil {
				//t.Errorf("Error reading from test client: %s", err.Error())
				return
			}

			if string(b) != responseText {
				//t.Errorf("Invalid response from client. Got '%s' expected '%s'", string(b), responseText)
				return
			}

			d <- nil
		}()
	}
}

func runLoadBalancer(d chan error) {
	dev := "eth0"
	if runtime.GOOS == "darwin" {
		dev = "en0"
	}

	lb := LoadBalancer{
		Name: "Test",
		IP:   net.IPv4(127, 0, 0, 1),
		Ports: []PortMap{
			PortMap{
				Src: 32144,
				Dst: 32145,
			},
		},
		Subnet:       net.IPv4(255, 255, 255, 255),
		Device:       dev,
		PollInterval: time.Second,
		Backends: Backends{
			&Backend{
				Name:    "TestBackend",
				IP:      net.IPv4(127, 0, 0, 1),
				Timeout: time.Second,
			},
		},
	}

	l := make(chan *LoadBalancer)
	lb.Run(l)
	d <- nil
}

func TestBackend(t *testing.T) {
	backendServerDone := make(chan error)
	go runBackendServer(backendServerDone, t)

	loadbalancerDone := make(chan error)
	go runLoadBalancer(loadbalancerDone)

	// Wait for loadbalancer to be listening...
	time.Sleep(time.Second)

	testConn, err := net.Dial("tcp", "127.0.0.1:32144")

	if err != nil {
		t.Errorf("Error connecting to loadbalancer: %s", err.Error())
		return
	}

	b := make([]byte, len(initialTextBytes))
	n, err := testConn.Read(b)

	if err != nil {
		t.Errorf("Error reading from loadbalancer: %s", err.Error())
		return
	}

	if n != len(initialTextBytes) {
		t.Errorf("Invalid message length from loadbalancer. Got %d expected %d", n, len([]byte(initialText)))
		return
	}

	decoded := string(b)

	if decoded != initialText {
		t.Errorf("Invalid message from loadbalancer. Got '%s' expected '%s'", decoded, initialText)
		return
	}

	n, err = testConn.Write(responseTextBytes)

	if err != nil {
		t.Errorf("Error writing to loadbalancer: %s", err.Error())
		return
	}

	if n != len(responseTextBytes) {
		t.Errorf("Failed to write all bytes to loadbalancer. Wrote %d expected %d", n, len([]byte(responseText)))
		return
	}

	select {
	case e := <-loadbalancerDone:
		if e != nil {
			t.Errorf("Error with load balancer: %s", e.Error())
		}
	case e := <-backendServerDone:
		if e != nil {
			t.Errorf("Error with backend server: %s", e.Error())
		}
		break
	}

	return
}
