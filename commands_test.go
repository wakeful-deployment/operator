package main

import (
	"errors"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/service"
	"github.com/wakeful-deployment/operator/test"
	"io/ioutil"
	"testing"
)

func TestValidLoadBootStateFromFile(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	f, err := ioutil.TempFile("", "operator-json")

	if err != nil {
		t.Fatal("Couldn't create a tmp file for this test")
	}

	_, err = f.WriteString(`{
	"metadata": {
		"foo": "bar"
	},
	"services": {
		"statsite": {
		"image": "wakeful/wake-statsite:latest",
		"ports": [{
			"incoming": 8125,
			"outgoing": 8125,
			"udp": true
		}],
		"env": {},
		"restart": "always",
		"tags": ["statsd", "udp"],
		"checks": []
	  }
	}
}`)

	if err != nil {
		t.Fatal("Couldn't write to the tmp file")
	}

	state := LoadBootStateFromFile(f.Name())

	m := state.Metadata
	lenM := len(m)

	if lenM != 1 {
		t.Errorf("Expected length of metadata to be 1, but got %d", lenM)
	}

	fooValue := m["foo"]
	if fooValue != "bar" {
		t.Errorf("Expected [foo] to be bar, but got %s", fooValue)
	}

	lenServices := len(state.Services)

	if lenServices != 1 {
		t.Fatalf("Expected 1 service, but got %d", lenServices)
	}

	s := state.Services["statsite"]

	if s.Name != "statsite" {
		t.Errorf("Expected service name to be statsite, but got %s", s.Name)
	}

	expectedImage := "wakeful/wake-statsite:latest"
	if s.Image != expectedImage {
		t.Errorf("Expected image to be %s, but got %s", expectedImage, s.Image)
	}
}

func TestInvalidLoadBootStateFromFile(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	f, err := ioutil.TempFile("", "operator-json")

	if err != nil {
		t.Fatal("Couldn't create a tmp file for this test")
	}

	_, err = f.WriteString(`"foo": "bar"}`)

	if err != nil {
		t.Fatal("Couldn't write to the tmp file")
	}

	state := LoadBootStateFromFile(f.Name())

	if state != nil {
		t.Fatal("State should be nil since we errored")
	}

	if !global.Machine.IsCurrently(global.ConfigFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.ConfigFailed, global.Machine.CurrentState)
	}
}

func TestBootWithStartAndRegister(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var startedContainers []string
	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
		RunResponse: func(c container.Container) error {
			startedContainers = append(startedContainers, c.Name)
			return nil
		},
		StopResponse: func(c container.Container) error { return nil },
	}

	var registeredServices []string
	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) { return "", nil },
		RegisterResponse: func(s service.Service) error {
			registeredServices = append(registeredServices, s.Name)
			return nil
		},
		DeregisterResponse:   func(s service.Service) error { return nil },
		DetectResponse:       func() error { return nil },
		PostMetadataResponse: func() error { return nil },
		ConsulHostResponse:   func() string { return "127.0.0.1" },
	}

	state := &State{Services: map[string]*service.Service{"foo": &service.Service{}}}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.Booted) {
		t.Errorf("Expected machine to be %s but was %v", global.Booted, global.Machine.CurrentState)
	}

	if len(startedContainers) != 1 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 1, len(startedContainers))
	}

	if len(registeredServices) != 1 {
		t.Errorf("Expected consul deregister to be called %d times but was called %d times", 1, len(registeredServices))
	}
}

func TestBootWithStopAndDeregister(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var stoppedContainers []string
	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return `921582f62758 consul`, nil },
		RunResponse:               func(c container.Container) error { return nil },
		StopResponse: func(c container.Container) error {
			stoppedContainers = append(stoppedContainers, c.Name)
			return nil
		},
	}

	var deregisteredServices []string
	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) {
			str := `{"consul":{"ID":"consul","Service":"consul","Tags":[],"Address":"","Port":8300}}`
			return str, nil
		},
		RegisterResponse: func(s service.Service) error { return nil },
		DeregisterResponse: func(s service.Service) error {
			deregisteredServices = append(deregisteredServices, s.Name)
			return nil
		},
		DetectResponse:       func() error { return nil },
		PostMetadataResponse: func() error { return nil },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.Booted) {
		t.Errorf("Expected machine to be %s but was %v", global.Booted, global.Machine.CurrentState)
	}

	if len(stoppedContainers) != 1 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 1, len(stoppedContainers))
	}

	if len(deregisteredServices) != 1 {
		t.Errorf("Expected consul deregister to be called %d times but was called %d times", 1, len(deregisteredServices))
	}
}

func TestBootFailedConsulCheck(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) { return "", nil },
		DetectResponse:             func() error { return errors.New("Not Detected") },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.ConsulFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.ConsulFailed, global.Machine.CurrentState)
	}
}

func TestBootFailedRunningContainersRequest(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", errors.New("Socket exploded") },
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) { return "", nil },
		DetectResponse:             func() error { return nil },
		PostMetadataResponse:       func() error { return nil },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.FetchingNodeStateFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.FetchingNodeStateFailed, global.Machine.CurrentState)
	}
}

func TestBootFailedRegisteredServicesRequest(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) { return "", errors.New("Socket exploded") },
		DetectResponse:             func() error { return nil },
		PostMetadataResponse:       func() error { return nil },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.FetchingNodeStateFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.FetchingNodeStateFailed, global.Machine.CurrentState)
	}
}

func TestBootFailedMetadataPost(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
	}

	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) { return "", nil },
		DetectResponse:             func() error { return nil },
		PostMetadataResponse:       func() error { return errors.New("metadata request failed") },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.PostingMetadataFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.PostingMetadataFailed, global.Machine.CurrentState)
	}
}

func TestSuccessfulTickWithStart(t *testing.T) {
	global.Machine.ForceTransition(global.Booted, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	var startedContainers []string
	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
		RunResponse: func(c container.Container) error {
			startedContainers = append(startedContainers, c.Name)
			return nil
		},
		StopResponse: func(c container.Container) error { return nil },
	}

	var registeredServices []string
	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) {
			return `{"consul":{"ID":"consul","Service":"consul","Tags":[],"Address":"","Port":8300},"statsite":{"ID":"statsite","Service":"statsite","Tags":null,"Address":"10.1.0.9","Port":0}}`, nil
		},
		RegisterResponse: func(s service.Service) error {
			registeredServices = append(registeredServices, s.Name)
			return nil
		},
		DeregisterResponse:   func(s service.Service) error { return nil },
		DetectResponse:       func() error { return nil },
		PostMetadataResponse: func() error { return nil },
		ConsulHostResponse:   func() string { return "127.0.0.1" },
		GetDirectoryStateResponse: func() (*consul.DirectoryState, error) {
			return &consul.DirectoryState{KVs: []consul.KV{consul.KV{Key: "_wakeful/nodes/981eb8e33da95184/apps/proxy", Value: "eyJpbWFnZSI6InBsdW0vd2FrZS1wcm94eTpsYXRlc3QiLCJ0YWdzIjpbXX0="}}}, nil
		},
	}

	bootState := &State{}
	directoryState := GetState(consulClient, "abcefg", 1, "5m")

	Tick(dockerClient, consulClient, bootState, directoryState)

	if !global.Machine.IsCurrently(global.Running) {
		t.Errorf("Expected machine to be %s but was %v", global.Booted, global.Machine.CurrentState)
	}

	if len(startedContainers) != 1 {
		t.Errorf("Expected docker start to be called %d times but was called %d times", 1, len(startedContainers))
	}

	if len(registeredServices) != 1 {
		t.Errorf("Expected to register %d service but %d were registered", 1, len(startedContainers))
	}
}
