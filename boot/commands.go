package boot

import (
	"fmt"
	"github.com/wakeful-deployment/operator/directory"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/node"
	"time"
)

func DetectOrBootConsul(state *State) error {
	// TODO
	return nil
}

func LoadStateFromFile(path string) *State {
	state, err := ReadStateFromConfigFile(path)

	if err != nil {
		global.Transition(global.NewState(global.ConfigFailed, err))
		return nil
	}

	return state
}

func Boot(state *State) {
	var err error

	global.Transition(global.NewState(global.Booting, nil))

	err = DetectOrBootConsul(state)

	if err != nil {
		global.Transition(global.NewState(global.ConsulFailed, err))
		return
	}

	currentNodeState, err := node.CurrentState()

	if err != nil {
		global.Transition(global.NewState(global.NodeStateFailed, err))
		return
	}

	err = Normalize(state, currentNodeState)

	if err != nil {
		global.Transition(global.NewState(global.NormalizeFailed, err))
		return
	}

	global.Transition(global.NewState(global.Booted, nil))
}

func MergeStateAndNormalize(currentState *State, directoryState *directory.State) error {
	desiredState, err := MergeState(currentState, directoryState)

	if err != nil {
		return err
	}

	currentNodeState, err := node.CurrentState()

	if err != nil {
		return err
	}

	err = Normalize(desiredState, currentNodeState)

	if err != nil {
		return err
	}

	return nil
}

func Once(currentState *State) error {
	directoryStateUrl := directory.StateURL{Wait: "0s"}
	directoryState, err := directory.GetState(directoryStateUrl.String())

	if err != nil {
		return err
	}

	err = MergeStateAndNormalize(currentState, directoryState)

	if err != nil {
		return err
	}

	return nil
}

func Loop(currentState *State, wait string) {
	directoryStateUrl := directory.StateURL{Wait: wait}

	for {
		directoryState, err := directory.GetState(directoryStateUrl.String()) // this will block for some time

		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		err = MergeStateAndNormalize(currentState, directoryState)

		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		directoryStateUrl.Index = directoryState.Index // for next iteration

		time.Sleep(time.Second)
	}
}
