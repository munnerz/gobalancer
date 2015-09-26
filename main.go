package main

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/munnerz/loadbalancer/tcp"
)

func main() {
	t := tcp.LoadBalancer{
		Name: "Test",
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 2000,
		Backends: []*tcp.Backend{
			&tcp.Backend{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 10000,
			},
			&tcp.Backend{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 10001,
			},
			&tcp.Backend{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 10002,
			},
			&tcp.Backend{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 10003,
			},
		},
		PollInterval: time.Millisecond * 100,
	}

	log.SetLevel(log.DebugLevel)

	t.Run()
}
