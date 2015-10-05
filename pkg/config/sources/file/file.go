package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/munnerz/gobalancer/pkg/api"
	"github.com/munnerz/gobalancer/pkg/config"
)

const (
	name = "file"
)

type File struct {
	filename string

	writeLock sync.Mutex
}

// GetConfig reads a JSON formatted configuration from disk
func (f *File) GetConfig() (*api.Config, error) {
	data, err := ioutil.ReadFile(f.filename)

	if err != nil {
		return nil, err
	}

	config := new(api.Config)

	err = json.Unmarshal(data, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves a Config object to a file on disk
func (f *File) SaveConfig(c *api.Config) error {
	data, err := json.Marshal(c)

	if err != nil {
		return err
	}

	f.writeLock.Lock()
	defer f.writeLock.Unlock()

	return ioutil.WriteFile(f.filename, data, os.ModePerm)
}

func NewFileStorage(params ...func(config.Storage) error) (config.Storage, error) {
	file := new(File)

	for _, p := range params {
		err := p(file)

		if err != nil {
			return file, err
		}
	}

	return file, nil
}

func SetFilename(c string) func(config.Storage) error {
	return func(p config.Storage) error {
		if a, ok := p.(*File); ok {
			a.filename = c
			return nil
		}
		return fmt.Errorf("SetFilename needs a *File instance, incorrect type passed")
	}
}

func init() {
	config.AddType(name, NewFileStorage)
}
