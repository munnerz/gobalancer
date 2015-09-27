package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/munnerz/gobalancer/config"
	"github.com/munnerz/gobalancer/logging"
	"github.com/munnerz/gobalancer/tcp"
)

var (
	configFile = flag.String("config", "rules.json", "input configuration JSON")
	executed   map[string]tcp.LoadBalancer
)

func init() {
	executed = make(map[string]tcp.LoadBalancer)
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
				if _, ok := executed[b.Name]; ok {
					continue
				}
				logging.Log("core", log.Infof, "Starting LoadBalancer '%s' (%s:%d)", b.Name, b.IP, b.Port)
				go b.Run(done)
				executed[b.Name] = b
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

	go execute(storage)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-signalChan

	logging.Log("core", log.Infof, "Closing load balancers...")

	for _, lb := range executed {
		err := lb.Stop()

		if err != nil {
			logging.Log("core", log.Errorf, "Error closing load balancer '%s': %s", lb.Name, err.Error())
		}
	}
}
