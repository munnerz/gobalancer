package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/munnerz/gobalancer/config"
	"github.com/munnerz/gobalancer/logging"
	"github.com/munnerz/gobalancer/tcp"
)

var (
	configFile   = flag.String("config", "rules.json", "input configuration JSON")
	executed     map[string]*tcp.LoadBalancer
	executedLock sync.RWMutex
)

func init() {
	executed = make(map[string]*tcp.LoadBalancer)
}

func execute(storage config.Storage) {
	in := make(chan *tcp.LoadBalancer)

	go func() {
		for {
			c, err := storage.GetConfig()
			if err != nil {
				continue
			}
			for _, b := range c.Loadbalancers.TCP {
				if _, ok := executed[b.Name]; ok {
					continue
				}
				logging.Log("core", log.Infof, "Starting LoadBalancer '%s' (%s:%v)", b.Name, b.IP, b.Ports)
				go b.Run(in)
				executed[b.Name] = b
			}
			time.Sleep(time.Second * 5)
		}
	}()

	for {
		lb := <-in

		DeleteLB(lb)
	}
}

func DeleteLB(l *tcp.LoadBalancer) {
	executedLock.Lock()
	defer executedLock.Unlock()
	delete(executed, l.Name)
}

func main() {
	flag.Parse()

	log.SetLevel(log.InfoLevel)

	storage := config.NewFileStorage(*configFile)

	go execute(storage)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-signalChan

	logging.Log("core", log.Infof, "Closing load balancers...")

	executedLock.Lock()
	defer executedLock.Unlock()

	for _, lb := range executed {
		logging.Log("core", log.Debugf, "Closing load balancer: %s", lb.Name)
		lb.Stop()
	}

}
