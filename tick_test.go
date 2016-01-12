package main

import (
	"errors"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/service"
	"github.com/wakeful-deployment/operator/test"
	"testing"
)

func TestSuccessfulTickWithStart(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var startedContainers []string
	var stoppedContainers []string
	dockerClient := dockerClient(&startedContainers, &stoppedContainers)
	dockerClient.RunningContainersResponse = func() (string, error) {
		return `
		operator plum/wake-operator:c60758244
		consul plum/wake-consul-agent:latest
		statsite plum/wake-statsite:latest
		`, nil
	}

	var registeredServices []string
	var deregisteredServices []string
	consulClient := consulClient(&registeredServices, &deregisteredServices)
	consulClient.RegisteredServicesResponse = func() (string, error) {
		return `{"consul":{"ID":"consul","Service":"consul","Tags":[],"Address":"","Port":8300},"statsite":{"ID":"statsite","Service":"statsite","Tags":null,"Address":"10.1.0.9","Port":0}}`, nil
	}

	proxyKV := consul.KV{Key: "_wakeful/nodes/981eb8e33da95184/apps/proxy", Value: "eyJpbWFnZSI6InBsdW0vd2FrZS1wcm94eTpsYXRlc3QiLCJ0YWdzIjpbXX0="}
	directoryState := &consul.DirectoryState{KVs: []consul.KV{proxyKV}}

	Tick(dockerClient, consulClient, bootState(), directoryState)

	if !global.Machine.IsCurrently(global.Running) {
		t.Errorf("Expected machine to be %s but was %v", global.Running, global.Machine.CurrentState)
	}

	if len(startedContainers) != 1 {
		t.Errorf("Expected docker start to be called %d times but was called %d times", 1, len(startedContainers))
	}

	if len(stoppedContainers) != 0 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 0, len(stoppedContainers))
	}

	if len(registeredServices) != 1 {
		t.Errorf("Expected to register %d services but %d were registered", 1, len(registeredServices))
	}

	if len(deregisteredServices) != 0 {
		t.Errorf("Expected to deregister %d services but %d were deregistered", 0, len(deregisteredServices))
	}
}

func TestSuccessfulTickWithStop(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var startedContainers []string
	var stoppedContainers []string
	dockerClient := dockerClient(&startedContainers, &stoppedContainers)
	dockerClient.RunningContainersResponse = func() (string, error) {
		return `
operator plum/wake-operator:c60758244
consul plum/wake-consul-agent:latest
statsite plum/wake-statsite:latest
proxy plum/wake-proxy:latest
		`, nil
	}

	var registeredServices []string
	var deregisteredServices []string
	consulClient := consulClient(&registeredServices, &deregisteredServices)
	consulClient.RegisteredServicesResponse = func() (string, error) {
		return `{"consul":{"ID":"consul","Service":"consul","Tags":[],"Address":"","Port":8300},"statsite":{"ID":"statsite","Service":"statsite","Tags":null,"Address":"10.1.0.9","Port":0}, "proxy":{"ID":"proxy","Service":"proxy","Tags":[],"Address":"","Port":8000}}`, nil
	}

	directoryState := &consul.DirectoryState{}

	Tick(dockerClient, consulClient, bootState(), directoryState)

	if !global.Machine.IsCurrently(global.Running) {
		t.Errorf("Expected machine to be %s but was %v", global.Running, global.Machine.CurrentState)
	}

	if len(startedContainers) != 0 {
		t.Errorf("Expected docker start to be called %d times but was called %d times", 0, len(startedContainers))
	}

	if len(stoppedContainers) != 1 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 1, len(stoppedContainers))
	}

	if len(registeredServices) != 0 {
		t.Errorf("Expected to register %d services but %d were registered", 0, len(registeredServices))
	}

	if len(deregisteredServices) != 1 {
		t.Errorf("Expected to deregister %d services but %d were deregistered", 1, len(deregisteredServices))
	}
}

func TestFailedTickDockerFailed(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var startedContainers []string
	var stoppedContainers []string
	dockerClient := dockerClient(&startedContainers, &stoppedContainers)
	dockerClient.RunningContainersResponse = func() (string, error) { return "", errors.New("Docker failed") }

	var registeredServices []string
	var deregisteredServices []string
	consulClient := consulClient(&registeredServices, &deregisteredServices)
	consulClient.RegisteredServicesResponse = func() (string, error) { return "", nil }

	Tick(dockerClient, consulClient, bootState(), &consul.DirectoryState{})

	if !global.Machine.IsCurrently(global.FetchingNodeStateFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.FetchingNodeStateFailed, global.Machine.CurrentState)
	}

	if len(startedContainers) != 0 {
		t.Errorf("Expected docker start to be called %d times but was called %d times", 0, len(startedContainers))
	}

	if len(stoppedContainers) != 0 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 0, len(stoppedContainers))
	}

	if len(registeredServices) != 0 {
		t.Errorf("Expected to register %d services but %d were registered", 0, len(registeredServices))
	}

	if len(deregisteredServices) != 0 {
		t.Errorf("Expected to deregister %d services but %d were deregistered", 0, len(deregisteredServices))
	}
}

func TestFailedTickConsulFailed(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var startedContainers []string
	var stoppedContainers []string
	dockerClient := dockerClient(&startedContainers, &stoppedContainers)
	dockerClient.RunningContainersResponse = func() (string, error) { return "", nil }

	var registeredServices []string
	var deregisteredServices []string
	consulClient := consulClient(&registeredServices, &deregisteredServices)
	consulClient.RegisteredServicesResponse = func() (string, error) { return "", errors.New("Consul Failed") }

	Tick(dockerClient, consulClient, bootState(), &consul.DirectoryState{})

	if !global.Machine.IsCurrently(global.FetchingNodeStateFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.FetchingNodeStateFailed, global.Machine.CurrentState)
	}

	if len(startedContainers) != 0 {
		t.Errorf("Expected docker start to be called %d times but was called %d times", 0, len(startedContainers))
	}

	if len(stoppedContainers) != 0 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 0, len(stoppedContainers))
	}

	if len(registeredServices) != 0 {
		t.Errorf("Expected to register %d services but %d were registered", 0, len(registeredServices))
	}

	if len(deregisteredServices) != 0 {
		t.Errorf("Expected to deregister %d services but %d were deregistered", 0, len(deregisteredServices))
	}
}

func TestSuccessfulGetDirectoryState(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	index := 123
	var registeredServices []string
	var deregisteredServices []string
	consulClient := consulClient(&registeredServices, &deregisteredServices)
	consulClient.GetDirectoryStateResponse = func() (*consul.DirectoryState, error) {
		return &consul.DirectoryState{Index: index}, nil
	}

	directoryState := GetDirectoryState(consulClient, "abc123", 0, "5m")

	if !global.Machine.IsCurrently(global.Booted) {
		t.Errorf("Expected machine to be %s but was %v", global.Booted, global.Machine.CurrentState)
	}

	if directoryState.Index != index {
		t.Errorf("expected index to be %d, but was %d", index, directoryState.Index)
	}
}

func TestFailedGetDirectoryState(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var registeredServices []string
	var deregisteredServices []string
	consulClient := consulClient(&registeredServices, &deregisteredServices)
	consulClient.GetDirectoryStateResponse = func() (*consul.DirectoryState, error) {
		return nil, errors.New("Fetching directory state failed")
	}

	directoryState := GetDirectoryState(consulClient, "abc123", 0, "5m")

	if !global.Machine.IsCurrently(global.FetchingDirectoryStateFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.FetchingDirectoryStateFailed, global.Machine.CurrentState)
	}

	if directoryState != nil {
		t.Errorf("expected directory state to be nil but was not")
	}
}

// Helpers

func bootState() *State {
	return &State{Services: map[string]*service.Service{"consul": &service.Service{Name: "consul"}, "statsite": &service.Service{Name: "statsite"}}}
}

func dockerClient(startedContainers *[]string, stoppedContainers *[]string) test.DockerClient {
	return test.DockerClient{
		RunResponse: func(c container.Container) error {
			*startedContainers = append(*startedContainers, c.Name)
			return nil
		},
		StopResponse: func(c container.Container) error {
			*stoppedContainers = append(*stoppedContainers, c.Name)
			return nil
		},
	}
}

func consulClient(registeredServices *[]string, deregisteredServices *[]string) test.ConsulClient {
	return test.ConsulClient{
		RegisterResponse: func(s service.Service) error {
			*registeredServices = append(*registeredServices, s.Name)
			return nil
		},
		DeregisterResponse: func(s service.Service) error {
			*deregisteredServices = append(*deregisteredServices, s.Name)
			return nil
		},
		PostMetadataResponse: func() error { return nil },
		ConsulHostResponse:   func() string { return "127.0.0.1" },
	}
}
