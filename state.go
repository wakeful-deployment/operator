package main

import (
	"bytes"
	"encoding/json"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/service"
	"io/ioutil"
)

type State struct {
	Metadata   map[string]string           `json:"metadata"`
	Services   map[string]*service.Service `json:"services"`
	NodeName   string                      `json:"node"`
	ConsulHost string                      `json:"consul"`
	ShouldLoop bool                        `json:"loop"`
	Wait       string                      `json:"wait"`
}

func ReadStateFromConfigFile(path string) (*State, error) {
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	state := &State{}
	jsonErr := json.NewDecoder(bytes.NewReader(contents)).Decode(state)

	if jsonErr != nil {
		return nil, jsonErr
	}

	for name, s := range state.Services {
		s.Name = name
	}

	return state, nil
}

func MergeStates(bootState *State, directoryState *consul.DirectoryState) (*State, error) {
	newState := &State{}
	*newState = *bootState // clone

	// The clone trick above is only a shallow clone.
	// Since we want to modify the services of newState, we must
	// first explicitly copy it over from bootState to ensure we don't
	// end up modifying bootState as well
	newState.Services = make(map[string]*service.Service)
	for k, v := range bootState.Services {
		newState.Services[k] = v
	}

	directoryServices, err := directoryState.Services()

	if err != nil {
		return nil, err
	}

	for _, s := range directoryServices {
		newState.Services[s.Name] = s
	}

	return newState, nil
}
