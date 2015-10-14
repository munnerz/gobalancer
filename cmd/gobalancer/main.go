package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/munnerz/gobalancer/pkg/addressing"
	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/config"
	"github.com/munnerz/gobalancer/pkg/config/sources/file"
	svc "github.com/munnerz/gobalancer/pkg/services"
)

var (
	configFile = flag.String("config", "config.json", "JSON formatted config file")

	ipPool        *addressing.IPPool
	configStorage config.Storage
	conf          *api.Config

	services    []*svc.Service
	serviceChan = make(chan svc.Error)
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
		services = append(services, svc.NewService(s, serviceChan))
		go services[i].Run()
	}

	// Wait for messages from different components...
	for {
		select {
		case e := <-serviceChan:
			log.Errorf("Error from service '%s': %s", e.Service.Name, e.Error())
		}
	}
}
