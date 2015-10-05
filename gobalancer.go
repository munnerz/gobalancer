package gobalancer

import (
	"fmt"
	"log"

	"github.com/munnerz/gobalancer/pkg/addressing"
	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/config"
)

var (
	ErrServiceAlreadyRegistered = fmt.Errorf("Service with that name already exists")
	ErrServiceNotFound          = fmt.Errorf("Service with given name could not be found")
)

type GoBalancer struct {
	configStorage config.Storage

	config    *api.Config
	allocator *addressing.Allocator
	services  map[string]*api.Service
}

func (g *GoBalancer) CreateService(s *api.Service) error {
	if _, ok := g.services[s.Name]; ok {
		return ErrServiceAlreadyRegistered
	}

	if s.IP == nil {
		ip, err := g.allocator.AllocateIP()

		if err != nil {
			return err
		}

		s.IP = ip

		g.configStorage.SaveConfig(g.config)
	}

	g.services[s.Name] = s

	return nil
}

func (g *GoBalancer) UpdateService(s *api.Service) error {
	if _, ok := g.services[s.Name]; !ok {
		return ErrServiceNotFound
	}

	return nil
}

func (g *GoBalancer) Run() error {
	for _, svc := range g.config.Services {
		err := g.CreateService(svc)

		if err != nil {
			// TODO: Make this catch and don't crash here?
			return err
		}
	}

	log.Printf("Loaded %d services. Launching loadbalancers...", len(g.services))

	for _, svc := range g.services {

	}

	return nil
}

func NewGoBalancer(c config.Storage) (*GoBalancer, error) {
	conf, err := c.GetConfig()

	if err != nil {
		return nil, err
	}

	alloc := addressing.NewAllocator(conf.Allocator)

	return &GoBalancer{
		configStorage: c,
		config:        conf,
		allocator:     alloc,
		services:      make(map[string]*api.Service),
	}, nil
}
