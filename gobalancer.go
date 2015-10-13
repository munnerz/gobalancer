package gobalancer

import (
	"fmt"
	"log"

	"github.com/munnerz/gobalancer/pkg/addressing"
	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/config"
	"github.com/munnerz/gobalancer/pkg/loadbalancers"
	_ "github.com/munnerz/gobalancer/pkg/loadbalancers/types"
)

var (
	ErrServiceAlreadyRegistered = fmt.Errorf("Service with that name already exists")
	ErrServiceNotFound          = fmt.Errorf("Service with given name could not be found")
)

type GoBalancer struct {
	configStorage config.Storage

	config   *api.Config
	pool     *addressing.IPPool
	services map[string]service
}

type service struct {
	*api.Service

	loadbalancers map[string]loadbalancers.LoadBalancer
}

func (g *GoBalancer) CreateService(s *api.Service) error {
	if _, ok := g.services[s.Name]; ok {
		return ErrServiceAlreadyRegistered
	}

	if s.IP == nil {
		ip, err := g.pool.AllocateIP()

		if err != nil {
			return err
		}

		s.IP = ip

		g.configStorage.SaveConfig(g.config)
	}

	err := g.pool.RegisterIP(*s.IP)

	if err != nil {
		return err
	}

	g.services[s.Name] = service{
		Service:       s,
		loadbalancers: make(map[string]loadbalancers.LoadBalancer),
	}

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

	errorChan := make(chan error)

ServiceLoop:
	for _, s := range g.services {
		for _, p := range s.Ports {
			lb, err := loadbalancers.NewLoadBalancer(s.IP.IP, p.Src, p.Type, s.Backends)

			if err != nil {
				log.Printf("Error creating loadbalancer: %s", err.Error())
				continue ServiceLoop
			}

			s.loadbalancers[p.Name] = lb
		}

		for _, l := range s.loadbalancers {
			go loadbalancers.RunLoadBalancer(l, errorChan)
		}
	}

	// TODO: CHANGE THIS ASAP
	return <-errorChan
}

func NewGoBalancer(c config.Storage) (*GoBalancer, error) {
	conf, err := c.GetConfig()

	if err != nil {
		return nil, err
	}

	pool := addressing.NewIPPool(conf.IPPool)

	return &GoBalancer{
		configStorage: c,
		config:        conf,
		pool:          pool,
		services:      make(map[string]service),
	}, nil
}
