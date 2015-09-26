package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/munnerz/gobalancer/tcp"
)

var (
	configFile = flag.String("config", "rules.json", "input configuration JSON")
)

type configDefinition struct {
	Loadbalancers struct {
		TCP []tcp.LoadBalancer `json:"tcp"`
	} `json:"loadbalancers"`
}

func main() {
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	configData, err := ioutil.ReadFile(*configFile)

	if err != nil {
		log.Fatalf("Error opening config file '%s': %s", *configFile, err.Error())
	}

	config := configDefinition{}

	err = json.Unmarshal(configData, &config)

	if err != nil {
		log.Fatalf("Error parsing config file '%s': %s", *configFile, err.Error())
	}

	done := make(chan error)

	for _, b := range config.Loadbalancers.TCP {
		go b.Run(done)
	}

	running := len(config.Loadbalancers.TCP)

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
