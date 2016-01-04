package boot

import (
	"bytes"
	"encoding/json"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/node"
	"io/ioutil"
)

type Config struct {
	MetaData map[string]string       `json:"metadata"`
	Services map[string]node.Service `json:"services"`
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

func NewConfig(prevConfig *Config, desiredState *consul.State) (*Config, error) {
	config := &Config{}
	*config = *prevConfig // clone

	// actually put the state into the config

	return config, nil
}
