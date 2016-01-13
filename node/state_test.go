package node

import (
	"errors"
	"github.com/wakeful-deployment/operator/test"
	"testing"
)

func registeredServices() (string, error) {
	str := `{"consul":{"ID":"consul","Service":"consul","Tags":[],"Address":"","Port":8300}}`
	return str, nil
}

func erroredRegisteredServices() (string, error) {
	return "", errors.New("I don't care")
}

func runningContainers() (string, error) {
	str := `921582f62758 consul
405bab56d0c7 statsite
3e02f2aae498 operator`

	return str, nil
}

func erroredRunningContainers() (string, error) {
	return "", errors.New("I don't care")
}

func TestSuccessfulCurrentState(t *testing.T) {
	dockerClient := test.DockerClient{
		RunningContainersResponse: runningContainers,
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: registeredServices,
	}

	state, err := CurrentState(dockerClient, consulClient)

	if err != nil {
		t.Errorf("Got an error: %v", err)
	}

	lenContainers := len(state.Containers)
	if lenContainers != 3 {
		t.Errorf("Expected zero containers, but found %d", lenContainers)
	}

	lenServices := len(state.Services)
	if lenServices != 1 {
		t.Errorf("Expected zero containers, but found %d", lenServices)
	}
}

func TestConsulErrorCurrentState(t *testing.T) {
	dockerClient := test.DockerClient{
		RunningContainersResponse: runningContainers,
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: erroredRegisteredServices,
	}

	_, err := CurrentState(dockerClient, consulClient)

	if err == nil {
		t.Error("We expected an error, but got none")
	}
}

func TestDockerErrorCurrentState(t *testing.T) {
	dockerClient := test.DockerClient{
		RunningContainersResponse: erroredRunningContainers,
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: registeredServices,
	}

	_, err := CurrentState(dockerClient, consulClient)

	if err == nil {
		t.Error("We expected an error, but got none")
	}
}
