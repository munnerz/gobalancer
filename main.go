package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/munnerz/gobalancer/config"
	"github.com/munnerz/gobalancer/logging"
)

var (
	configFile = flag.String("config", "rules.json", "input configuration JSON")
)

func main() {
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	storage := config.NewFileStorage(*configFile)

	c, err := storage.GetConfig()

	if err != nil {
		logging.Log("core", log.Errorf, "Error loading config: %s", err.Error())
	}

	done := make(chan error)

	for _, b := range c.Loadbalancers.TCP {
		go b.Run(done)
	}

	running := len(c.Loadbalancers.TCP)

	for {
		if running == 0 {
			break
		}

		err := <-done

		running--

		if err != nil {
			log.Errorf("LoadBalancer crashed: %s", err.Error())
			continue
		}
	}
}
