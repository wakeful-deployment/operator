package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

type BootImage struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Ports []PortPair
	Env   map[string]string `json:"env"`
}

func (b BootImage) ToContainer() Container {
	return Container{Name: b.Name, Image: b.Image, Ports: b.Ports, Env: b.Env}
}

type Config struct {
	BootImages []BootImage `json:"boot_images"`
}

func NewConfig(configPath string) (*Config, error) {
	contents, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	config := &Config{}
	jsonErr := json.NewDecoder(bytes.NewReader(contents)).Decode(config)

	if jsonErr != nil {
		return nil, jsonErr
	}

	return config, nil
}

func bootstrapContainers(config *Config) []Container {
	var containers []Container
	for _, bootImage := range config.BootImages {
		containers = append(containers, bootImage.ToContainer())
	}

	return containers
}

func runBootstrapContainers(containers []Container, runningContainers []Container) {
	for _, container := range containers {
		containerAlreadyRunning := false
		for _, runningContainer := range runningContainers {
			if runningContainer.Name == container.Name {
				containerAlreadyRunning = true
			}
		}

		if !containerAlreadyRunning {
			container.Run()
		}
	}
}

func Bootstrap(configPath string) ([]Container, error) {
	config, err := NewConfig(configPath)

	if err != nil {
		return nil, err
	}

	containers := bootstrapContainers(config)
	runningContainers, err := RunningContainers()

	if err != nil {
		return nil, err
	}

	runBootstrapContainers(containers, runningContainers)

	return containers, nil
}
