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
	data, err := ioutil.ReadFile(f.filename)

	if err != nil {
		return nil, err
	}

	config := Config{}

	err = json.Unmarshal(data, &config)

	if err != nil {
		return nil, err
	}

	err = f.Memory.SaveConfig(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (f *File) SaveConfig(c *Config) error {
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

func (f *File) AddTCPLoadbalancers(t ...tcp.LoadBalancer) error {
	err := f.Memory.AddTCPLoadbalancers(t...)

	if err != nil {
		return err
	}

	cfg, err := f.Memory.GetConfig()

	if err != nil {
		return err
	}

	return f.SaveConfig(cfg)
}

func NewFileStorage(filename string) *File {
	return &File{
		filename: filename,
		Memory:   NewMemoryStorage(nil),
	}
}
