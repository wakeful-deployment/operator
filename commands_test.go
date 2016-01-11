package main

import (
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

func TestSuccessfulBootWithStartAndRegister(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	timesStartCalled := 0
	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
		RunResponse:               func(c container.Container) error { timesStartCalled += 1; return nil },
		StopResponse:              func(c container.Container) error { return nil },
	}

	timesRegisterCalled := 0
	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) { return "", nil },
		RegisterResponse:           func(s service.Service) error { timesRegisterCalled += 1; return nil },
		DeregisterResponse:         func(s service.Service) error { return nil },
		DetectResponse:             func() error { return nil },
		PostMetadataResponse:       func() error { return nil },
		ConsulHostResponse:         func() string { return "127.0.0.1" },
	}

	state := &State{Services: map[string]*service.Service{"foo": &service.Service{}}}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.Booted) {
		t.Errorf("Expected machine to be %s but was %v", global.ConfigFailed, global.Machine.CurrentState)
	}

	if timesStartCalled != 1 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 1, timesStartCalled)
	}

	if timesRegisterCalled != 1 {
		t.Errorf("Expected consul deregister to be called %d times but was called %d times", 1, timesRegisterCalled)
	}
}

func TestSuccessfulBootWithStopAndDeregister(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	timesStopCalled := 0
	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return `921582f62758 consul`, nil },
		RunResponse:               func(c container.Container) error { return nil },
		StopResponse:              func(c container.Container) error { timesStopCalled += 1; return nil },
	}

	timesDeregisterCalled := 0
	consulClient := test.ConsulClient{
		RegisteredServicesResponse: func() (string, error) {
			str := `{"consul":{"ID":"consul","Service":"consul","Tags":[],"Address":"","Port":8300}}`
			return str, nil
		},
		RegisterResponse:     func(s service.Service) error { return nil },
		DeregisterResponse:   func(s service.Service) error { timesDeregisterCalled += 1; return nil },
		DetectResponse:       func() error { return nil },
		PostMetadataResponse: func() error { return nil },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.Booted) {
		t.Errorf("Expected machine to be %s but was %v", global.ConfigFailed, global.Machine.CurrentState)
	}

	if timesStopCalled != 1 {
		t.Errorf("Expected docker stop to be called %d times but was called %d times", 1, timesStopCalled)
	}

	if timesDeregisterCalled != 1 {
		t.Errorf("Expected consul deregister to be called %d times but was called %d times", 1, timesDeregisterCalled)
	}
}
