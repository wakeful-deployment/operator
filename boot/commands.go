package boot

import (
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
		global.Transition(global.NewState(global.FetchingNodeStateFailed, err))
		return
	}

	err = Normalize(state, currentNodeState)

	if err != nil {
		global.Transition(global.NewState(global.NormalizingFailed, err))
		return
	}

	global.Transition(global.NewState(global.Booted, nil))
}

func Run(currentState *State, wait string) {
	if global.CurrentState.NotEqual(global.Running) && global.CurrentState.NotEqual(global.Booted) {
		global.Transition(global.NewState(global.AttemptingToRecover, global.CurrentState.Error))
	}

	directoryStateUrl := directory.StateURL{Wait: wait}

	directoryState, err := directory.GetState(directoryStateUrl.String()) // this will block for some time

	if err != nil {
		global.Transition(global.NewState(global.FetchingDirectoryStateFailed, err))
		return
	}

	desiredState, err := MergeState(currentState, directoryState)

	if err != nil {
		global.Transition(global.NewState(global.MergingStateFailed, err))
		return
	}

	currentNodeState, err := node.CurrentState()

	if err != nil {
		global.Transition(global.NewState(global.FetchingNodeStateFailed, err))
		return
	}

	err = Normalize(desiredState, currentNodeState)

	if err != nil {
		global.Transition(global.NewState(global.NormalizingFailed, err))
		return
	}

	global.Transition(global.NewState(global.Running, nil))

	directoryStateUrl.Index = directoryState.Index // for next iteration
}

func Once(currentState *State) {
	Run(currentState, "5s")
}

func Loop(currentState *State, wait string) {
	for {
		Run(currentState, wait)
		time.Sleep(time.Second)
	}
}
