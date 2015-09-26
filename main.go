package main

import (
	"flag"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/munnerz/gobalancer/config"
	"github.com/munnerz/gobalancer/logging"
)

var (
	configFile = flag.String("config", "rules.json", "input configuration JSON")
	executed   map[string]bool
)

func init() {
	executed = make(map[string]bool)
}

func execute(storage config.Storage) {
	done := make(chan error)

	go func() {
		for {
			c, err := storage.GetConfig()
			if err != nil {
				done <- err
				continue
			}
			for _, b := range c.Loadbalancers.TCP {
				if executed[b.Name] {
					continue
				}
				logging.Log("core", log.Infof, "Starting LoadBalancer '%s' (%s:%d)", b.Name, b.IP, b.Port)
				go b.Run(done)
				executed[b.Name] = true
			}
			time.Sleep(time.Second * 5)
		}
	}()

	for {
		err := <-done

		if err != nil {
			log.Errorf("Executor error: %s", err.Error())
			continue
		}
	}
}

func main() {
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	storage := config.NewFileStorage(*configFile)

	execute(storage)
}
