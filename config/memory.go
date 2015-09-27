package config

import (
	"fmt"

	"github.com/munnerz/gobalancer/tcp"
)

type Memory struct {
	*Config
}

func (m *Memory) GetConfig() (*Config, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("No config in memory")
	}
	return m.Config, nil
}

func (m *Memory) SaveConfig(c *Config) error {
	m.Config = c
	return nil
}

func (m *Memory) AddTCPLoadbalancers(t ...*tcp.LoadBalancer) error {
	if m.Config == nil {
		return fmt.Errorf("No config in memory")
	}

	m.Config.Loadbalancers.TCP = append(m.Config.Loadbalancers.TCP, t...)

	return nil
}

func NewMemoryStorage(c *Config) *Memory {
	return &Memory{
		Config: c,
	}
}
