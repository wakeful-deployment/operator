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
func Normalize(desiredState *State, currentNodeState *node.State) error {
	// always try to fix the containers before fixing the registrations

	var desiredContainers []container.Container

	for _, service := range desiredState.Services {
		desiredContainers = append(desiredContainers, service.Container(desiredState.NodeName))
	}

	err := docker.NormalizeContainers(desiredContainers, currentNodeState.Containers)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// then try to register everything correctly in consul

	var desiredServices []service.Service

	for _, s := range desiredState.Services {
		desiredServices = append(desiredServices, s)
	}

	err = consul.NormalizeServices(desiredServices, currentNodeState.Services)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
