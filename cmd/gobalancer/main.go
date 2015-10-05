package main

import (
	"flag"
	"log"

	"github.com/munnerz/gobalancer/pkg/config"
	"github.com/munnerz/gobalancer/pkg/config/sources/file"

	"github.com/munnerz/gobalancer"
)

var (
	configFile = flag.String("config", "config.json", "JSON formatted config file")

	gb *gobalancer.GoBalancer
)

func main() {
	flag.Parse()

	fileStorage, err := config.GetType("file")

	if err != nil {
		log.Fatalf(err.Error())
	}

	storage, err := fileStorage(file.SetFilename(*configFile))

	if err != nil {
		log.Fatalf("Error getting config storage module: %s", err.Error())
	}

	gb, err = gobalancer.NewGoBalancer(storage)

	if err != nil {
		log.Fatalf("Error initialising GoBalancer core: %s", err.Error())
	}

	err = gb.Run()

	if err != nil {
		log.Fatalf("Error running gobalancer: %s", err.Error())
	}
}
