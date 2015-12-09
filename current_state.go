package main

type CurrentState struct {
	containers []Container
	services   []ConsulService
}

func (c CurrentState) Containers() []Container {
	return c.containers
}

func (c CurrentState) Services() []ConsulService {
	return c.services
}

func NewCurrentState(servicesUrl string) (*CurrentState, error) {
	// TODO: This find all keys in namespace that differ.
	// We want to only find 'app' keys
	currentContainers, err := RunningContainers()

	if err != nil {
		return nil, err
	}

	currentServices, err := GetConsulServices(servicesUrl)

	if err != nil {
		return nil, err
	}

	currentState := CurrentState{containers: currentContainers, services: currentServices}
	return &currentState, nil
}
