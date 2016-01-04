package boot

import (
	"bytes"
	"encoding/json"
	"github.com/wakeful-deployment/operator/directory"
	"github.com/wakeful-deployment/operator/service"
	"io/ioutil"
)

type State struct {
	MetaData map[string]string          `json:"metadata"`
	Services map[string]service.Service `json:"services"`
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

func MergeState(currentState *State, directoryState *directory.State) (*State, error) {
	newState := &State{}
	*newState = *currentState // clone

	directoryServices, err := directoryState.Services()

	if err != nil {
		return nil, err
	}

	for _, s := range directoryServices {
		newState.Services[s.Name] = s
	}

	return newState, nil
}
