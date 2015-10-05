package config

import (
	"fmt"

	"github.com/munnerz/gobalancer/pkg/api"
)

var (
	types map[string]func(...func(Storage) error) (Storage, error)
)

func AddType(name string, f func(...func(Storage) error) (Storage, error)) {
	types[name] = f
}

func GetType(name string) (func(...func(Storage) error) (Storage, error), error) {
	if f, ok := types[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("Unknown config provider: %s", name)
}

type Storage interface {
	GetConfig() (*api.Config, error)
	SaveConfig(*api.Config) error
}

func init() {
	types = make(map[string]func(...func(Storage) error) (Storage, error))
}
