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

type IPPool struct {
	Object

	Device  string    `json:"service_net_dev"`
	Network net.IPNet `json:"service_ip_range"`
}

// Backend contains configuration for a single loadbalancer backend
type Backend struct {
	Object

	IP           net.IP        `json:"ip"`
	PollInterval time.Duration `json:"poll_interval"`
	PollTimeout  time.Duration `json:"poll_timeout"`
}

type Service struct {
	Object

	IP    *net.IPNet     `json:"ip,omitempty"`
	Ports []*ServicePort `json:"ports"`

	Backends []*Backend `json:"backends"`
}

type ServicePortType string

const (
	ServicePortTypeTCP ServicePortType = "tcp"
	ServicePortTypeUDP ServicePortType = "udp"
)

// PortMap represents a mapping between two ports
type ServicePort struct {
	Object

	Type ServicePortType `json:"type"`
	Src  int             `json:"src"`
	Dst  int             `json:"dst"`
}
