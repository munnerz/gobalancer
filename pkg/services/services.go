package services

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/loadbalancers"
	"github.com/munnerz/gobalancer/pkg/utils"
)

type Service struct {
	Name     string
	IP       net.IPNet
	PortMaps []*api.PortMap
	Backends []*api.Backend

	loadbalancerErrorChan chan utils.Error
	errorChan             chan utils.Error
	loadbalancers         map[string]loadbalancers.LoadBalancer
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
			s.errorChan <- utils.NewError(s, err)
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

		var err error

		switch e.Error {
		case loadbalancers.ErrLoadBalancerStopped:
			log.Debugf("Stopped loadbalancer: %s", e.Sender.(loadbalancers.LoadBalancer).Name())
		default:
			log.Errorf("Error from loadbalancer '%s': %s", e.Sender.(loadbalancers.LoadBalancer).Name(), e.Error)
			err = e.Error
		}

		delete(s.loadbalancers, e.Sender.(loadbalancers.LoadBalancer).Name())

		if err != nil {
			s.errorChan <- utils.NewError(s, err)
		}

		// If there are no loadbalancers left in this service, let's exit
		if len(s.loadbalancers) == 0 {
			return
		}
	}
}

func (s *Service) Stop() {
	for _, l := range s.loadbalancers {
		log.Debugf("Stopping loadbalancer: %s", l.Name())
		l.Stop()
	}
}

func NewService(s *api.Service, errorChan chan utils.Error) *Service {
	return &Service{
		Name:                  s.Name,
		IP:                    *s.IP,
		PortMaps:              s.Ports,
		Backends:              s.Backends,
		loadbalancerErrorChan: make(chan utils.Error),
		errorChan:             errorChan,
		loadbalancers:         make(map[string]loadbalancers.LoadBalancer),
	}
}
