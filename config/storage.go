package config

import (
	"github.com/munnerz/gobalancer/tcp"
)

type Storage interface {
	GetConfig() (*Config, error)
	SaveConfig(*Config) error
	AddTCPLoadbalancers(...tcp.LoadBalancer) error
}

type Config struct {
	Loadbalancers Loadbalancers `json:"loadbalancers"`
}

type Loadbalancers struct {
	TCP []tcp.LoadBalancer `json:"tcp"`
}
