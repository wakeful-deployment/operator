package boot

import (
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
)

type State struct {
	Containers []docker.Container
	Services   []consul.Service
}

func CurrentState() (*State, error) {
	containers, err := docker.RunningContainers()

	if err != nil {
		return nil, err
	}

	services, err := consul.RegisteredServices()

	if err != nil {
		return nil, err
	}

	currentState := State{Containers: containers, Services: services}

	return &currentState, nil
}
