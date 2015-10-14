package services

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/loadbalancers"
	_ "github.com/munnerz/gobalancer/pkg/loadbalancers/types"
)

type Service struct {
	Name     string
	IP       net.IPNet
	PortMaps []*api.PortMap
	Backends []*api.Backend

	loadbalancerErrorChan chan error
	errorChan             chan Error
	loadbalancers         map[string]loadbalancers.LoadBalancer
}

type Error struct {
	error
	Service *Service
}

func (s *Service) error(err error) Error {
	return Error{
		error:   err,
		Service: s,
	}
}

func (s *Service) Run() {
	for _, p := range s.PortMaps {
		log.Debugf("Creating loadbalancer: %s", p.Name)
		lb, err := loadbalancers.NewLoadBalancer(loadbalancers.LoadBalancerSpec{
			Name:     p.Name,
			IP:       s.IP.IP,
			PortMap:  p,
			Backends: s.Backends,
		})

		if err != nil {
			s.errorChan <- s.error(err)
			return
		}

		s.loadbalancers[p.Name] = lb
	}

	for _, l := range s.loadbalancers {
		log.Debugf("Running loadbalancer: %s", l.Name())
		go loadbalancers.RunLoadBalancer(l, s.loadbalancerErrorChan)
	}

	for {
		e := <-s.loadbalancerErrorChan
		// Handle this error.. Perhaps exit the whole service for now?
		s.errorChan <- s.error(e)
		return
	}
}

func (s *Service) Stop() {
	for _, l := range s.loadbalancers {
		log.Debugf("Stopping loadbalancer: %s", l.Name())
		l.Stop()
	}
}

func NewService(s *api.Service, errorChan chan Error) *Service {
	return &Service{
		Name:                  s.Name,
		IP:                    *s.IP,
		PortMaps:              s.Ports,
		Backends:              s.Backends,
		loadbalancerErrorChan: make(chan error),
		errorChan:             errorChan,
		loadbalancers:         make(map[string]loadbalancers.LoadBalancer),
	}
}
