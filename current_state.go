package main

import (
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
)

type CurrentState struct {
	containers []docker.Container
	services   []consul.ConsulService
}

func (c CurrentState) Containers() []docker.Container {
	return c.containers
}

func (c CurrentState) Services() []consul.ConsulService {
	return c.services
}

func NewCurrentState(servicesUrl string) (*CurrentState, error) {
	// TODO: This find all keys in namespace that differ.
	// We want to only find 'app' keys
	currentContainers, err := docker.RunningContainers()

	if err != nil {
		return nil, err
	}

	currentServices, err := consul.GetConsulServices(servicesUrl)

	if err != nil {
		return nil, err
	}

	currentState := CurrentState{containers: currentContainers, services: currentServices}
	return &currentState, nil
}
