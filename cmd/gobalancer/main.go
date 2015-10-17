package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/gobalancer/pkg/addressing"
	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/config"
	"github.com/munnerz/gobalancer/pkg/config/sources/file"
	_ "github.com/munnerz/gobalancer/pkg/loadbalancers/types"
	svc "github.com/munnerz/gobalancer/pkg/services"
	"github.com/munnerz/gobalancer/pkg/utils"
)

var (
	configFile = flag.String("config", "config.json", "JSON formatted config file")

	ipPool        *addressing.IPPool
	configStorage config.Storage
	conf          *api.Config

	services    []*svc.Service
	serviceChan = make(chan utils.Error)
)

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	// Get a config initialiser
	NewFileStorage, err := config.GetType("file")

	if err != nil {
		log.Fatalf(err.Error())
	}

	// Initialise config FileStorage
	cs, err := NewFileStorage(file.SetFilename(*configFile))

	if err != nil {
		log.Fatalf("Error getting config storage module: %s", err.Error())
	}

	configStorage = cs

	// Read configuration
	c, err := configStorage.GetConfig()

	if err != nil {
		log.Fatalf("Error getting config: %s", err)
	}

	conf = c

	// Initialise IP address pool to allocate from
	ipPool = addressing.NewIPPool(conf.IPPool)

	for i, s := range conf.Services {
		if s.IP == nil {
			ip, err := ipPool.AllocateIP()

			if err != nil {
				log.Errorf("Error allocating IP address for service '%s': %s", s.Name, err)
				continue
			}

			s.IP = ip

			configStorage.SaveConfig(conf)
		}

		log.Debugf("Registering: %s", s.IP)

		err := ipPool.RegisterIP(*s.IP)

		if err != nil {
			log.Errorf("Error registering IP address for service '%s': %s", s.Name, err)
			continue
		}

		services = append(services, svc.NewService(s, serviceChan))
		go services[i].Run()
	}

	// Wait for messages from different components...
	for {
		select {
		case e := <-serviceChan:
			log.Errorf("Error from service '%s': %s", e.Sender.(api.Object).Name, e.Error.Error())
		}
	}
}
