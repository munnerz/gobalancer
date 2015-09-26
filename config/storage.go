package config

import (
	"github.com/munnerz/gobalancer/tcp"
)

type Storage interface {
	GetConfig() (*Config, error)
	SaveConfig(Config) error
	SaveTCPLoadbalancers(...*tcp.LoadBalancer) error
}

type Config struct {
	Loadbalancers struct {
		TCP []*tcp.LoadBalancer `json:"tcp"`
	} `json:"loadbalancers"`
}
