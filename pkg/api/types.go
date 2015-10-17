package api

import (
	"net"
	"time"
)

type Config struct {
	IPPool   *IPPool    `json:"ip_pool"`
	Services []*Service `json:"services"`
}

// Object is a generic for any named item
type Object struct {
	Name string `json:"name"`
}

type IPRange struct {
	Start net.IP `json:"start"`
	Stop  net.IP `json:"stop"`
}

type IPPool struct {
	Object

	Device  string    `json:"service_net_dev"`
	Network net.IPNet `json:"service_net"`
	Range   IPRange   `json:"service_net_range"`
}

// Backend contains configuration for a single loadbalancer backend
type Backend struct {
	Object

	Hostname     string        `json:"ip"`
	PollInterval time.Duration `json:"poll_interval"`
	PollTimeout  time.Duration `json:"poll_timeout"`
}

type Service struct {
	Object

	IP    *net.IPNet `json:"ip,omitempty"`
	Ports []*PortMap `json:"ports"`

	Backends []*Backend `json:"backends"`
}

type PortMapProtocol string

const (
	PortMapProtocolTCP PortMapProtocol = "tcp"
	PortMapProtocolUDP PortMapProtocol = "udp"
)

// PortMap represents a mapping between two ports
type PortMap struct {
	Object

	Protocol PortMapProtocol `json:"protocol"`
	Src      int             `json:"src"`
	Dst      int             `json:"dst"`
}
