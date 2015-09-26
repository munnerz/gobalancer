package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/munnerz/gobalancer/tcp"
)

type File struct {
	filename string
	*Memory
}

func (f *File) GetConfig() (*Config, error) {
	c, err := f.Memory.GetConfig()

	if err == nil {
		return c, nil
	}

	data, err := ioutil.ReadFile(f.filename)

	if err != nil {
		return nil, err
	}

	config := Config{}

	err = json.Unmarshal(data, &config)

	if err != nil {
		return nil, err
	}

	err = f.Memory.SaveConfig(config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (f *File) SaveConfig(c Config) error {
	data, err := json.Marshal(c)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f.filename, data, os.ModePerm)

	if err != nil {
		return err
	}

	err = f.Memory.SaveConfig(c)

	if err != nil {
		return err
	}

	return nil
}

func (f *File) SaveTCPLoadbalancers(t ...*tcp.LoadBalancer) error {
	config, err := f.GetConfig()

	if err != nil {
		return err
	}

	if config.Loadbalancers.TCP == nil {
		config.Loadbalancers.TCP = t
		return nil
	}

	config.Loadbalancers.TCP = append(config.Loadbalancers.TCP, t...)
	return nil
}

func NewFileStorage(filename string) *File {
	return &File{
		filename: filename,
		Memory:   NewMemoryStorage(nil),
	}
}
