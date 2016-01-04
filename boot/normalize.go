package boot

import (
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/node"
)

// reconcile the desired config with the current state
func Normalize(config *Config, currentState *State) error {
	// always try to fix the containers before fixing the registrations

	var desiredContainers node.ContainerCollection

	for _, service := range config.Services {
		desiredContainers.Append(service.Container())
	}

	err := docker.NormalizeDockerContainers(desiredContainers, currentState.Containers)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// then try to register everything correctly in consul

	var desiredServices []node.ConsulService

	for _, service := range config.Services {
		desiredServices = append(desiredServices, service.ConsulService())
	}

	err = consul.NormalizeConsulServices(desiredServices, currentState.Services)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
