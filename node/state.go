package node

import (
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/service"
)

type State struct {
	Containers []container.Container
	Services   []service.Service
}

func CurrentState(dockerClient docker.Client, consulClient consul.Client) (*State, error) {
	containers, err := docker.RunningContainers(dockerClient)

	if err != nil {
		return nil, err
	}

	services, err := consul.RegisteredServices(consulClient)

	if err != nil {
		return nil, err
	}

	currentState := State{Containers: containers, Services: services}

	return &currentState, nil
}
