package boot

import (
	"bytes"
	"encoding/json"
	"github.com/wakeful-deployment/operator/directory"
	"github.com/wakeful-deployment/operator/service"
	"io/ioutil"
)

type Config struct {
	MetaData map[string]string          `json:"metadata"`
	Services map[string]service.Service `json:"services"`
}

func ReadConfigFile(configPath string) (*Config, error) {
	contents, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	config := &Config{}
	jsonErr := json.NewDecoder(bytes.NewReader(contents)).Decode(config)

	if jsonErr != nil {
		return nil, jsonErr
	}

	for name, s := range config.Services {
		s.Name = name
	}

	return config, nil
}

func NewConfig(prevConfig *Config, desiredState *directory.State) (*Config, error) {
	config := &Config{}
	*config = *prevConfig // clone

	desiredServices, err := desiredState.Services()

	if err != nil {
		return nil, err
	}

	for _, s := range desiredServices {
		config.Services = append(config.Services, s)
	}

	return config, nil
}
