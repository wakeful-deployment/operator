package main

import (
	"errors"
	"github.com/wakeful-deployment/operator/global"
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

func TestBootFailedConsulCheck(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
	}

	consulClient := test.ConsulClient{
		DetectResponse: func() error { return errors.New("Not Detected") },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.ConsulFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.ConsulFailed, global.Machine.CurrentState)
	}
}

func TestBootFailedMetadataPost(t *testing.T) {
	global.Machine.ForceTransition(global.Initial, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	dockerClient := test.DockerClient{
		RunningContainersResponse: func() (string, error) { return "", nil },
	}

	consulClient := test.ConsulClient{
		DetectResponse:       func() error { return nil },
		PostMetadataResponse: func() error { return errors.New("metadata request failed") },
	}

	state := &State{}

	Boot(dockerClient, consulClient, state)

	if !global.Machine.IsCurrently(global.PostingMetadataFailed) {
		t.Errorf("Expected machine to be %s but was %v", global.PostingMetadataFailed, global.Machine.CurrentState)
	}
}
