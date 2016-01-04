package boot

import (
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/node"
	"github.com/wakeful-deployment/operator/service"
)

// reconcile the desired config with the current state
func Normalize(config *Config, currentState *node.State) error {
	// always try to fix the containers before fixing the registrations

	var desiredContainers []container.Container

	for _, service := range config.Services {
		desiredContainers = append(desiredContainers, service.Container())
	}

	err := docker.NormalizeDockerContainers(desiredContainers, currentState.Containers)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// then try to register everything correctly in consul

	var desiredServices []service.Service

	for _, s := range config.Services {
		desiredServices = append(desiredServices, s)
	}

	err = consul.NormalizeServices(desiredServices, currentState.Services)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
