package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"
)

func tick(stateUrl *ConsulStateURL, servicesUrl *ConsulServicesURL, bootstrapContainers []Container) (int, error) {
	desiredState, err := NewConsulState(stateUrl.ToString())

	if err != nil {
		return 0, err
	}

	services, sErr := GetConsulServices(*servicesUrl)

	if sErr != nil {
		fmt.Sprintf("ERROR: error getting the registered consul services: %v", err)
	}

	normalizeDockerContainers(*desiredState, bootstrapContainers)
	normalizeConsulServices(*desiredState, services, stateUrl.Host)

	return desiredState.Index, nil
}

func loop(stateUrl *ConsulStateURL, servicesUrl *ConsulServicesURL, bootstrapContainers []Container) {
	for {
		newIndex, err := tick(stateUrl, servicesUrl, bootstrapContainers)
		stateUrl.Index = newIndex

		time.Sleep(time.Second)

		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

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

func config(configPath string) (*Config, error) {
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

func bootstrap(configPath string) ([]Container, error) {
	config, err := config(configPath)

	if err != nil {
		return nil, err
	}

	containers := bootstrapContainers(config)
	runningContainers, err := runningContainers()

	if err != nil {
		return nil, err
	}

	runBootstrapContainers(containers, runningContainers)

	return containers, nil
}

func main() {
	var (
		hostname   = flag.String("hostname", "", "The name of the host which is running operator")
		consulHost = flag.String("consulhost", "", "The name of the consul host")
		configPath = flag.String("configpath", "./operator.json", "The path to the operator.json. The default is the current directory of the binary.")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "5m", "The timeout for polling")
	)
	flag.Parse()

	if *hostname == "" && *consulHost == "" {
		panic("ERROR: Must provide hostname and consulhost flags")
	}

	bootstrapContainers, err := bootstrap(*configPath)

	if err != nil {
		panic(err)
	}

	stateUrl := ConsulStateURL{Host: *consulHost, Hostname: *hostname, Wait: *wait}
	servicesUrl := ConsulServicesURL{Host: *consulHost}

	if *shouldLoop {
		loop(&stateUrl, &servicesUrl, bootstrapContainers)
	} else {
		tick(&stateUrl, &servicesUrl, bootstrapContainers)
	}
}
