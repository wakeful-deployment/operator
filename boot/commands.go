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

func LoadBootStateFromFile(path string) *State {
	state, err := ReadStateFromConfigFile(path)

	if err != nil {
		global.Machine.Transition(global.ConfigFailed, err)
		return nil
	}

	return state
}

func Boot(state *State) {
	var err error

	global.Machine.Transition(global.Booting, nil)

	err = DetectOrBootConsul(state)

	if err != nil {
		global.Machine.Transition(global.ConsulFailed, err)
		return
	}

	currentNodeState, err := node.CurrentState()

	if err != nil {
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		return
	}

	err = Normalize(state, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		return
	}

	global.Machine.Transition(global.Booted, nil)
}

func Run(bootState *State, wait string) {
	if !global.Machine.IsCurrently(global.Running) && !global.Machine.IsCurrently(global.Booted) {
		global.Machine.Transition(global.AttemptingToRecover, global.Machine.CurrentState.Error)
	}

	directoryStateUrl := directory.StateURL{Wait: wait}

	directoryState, err := directory.GetState(directoryStateUrl.String()) // this will block for some time

	if err != nil {
		global.Machine.Transition(global.FetchingDirectoryStateFailed, err)
		return
	}

	desiredState, err := MergeStates(bootState, directoryState)

	if err != nil {
		global.Machine.Transition(global.MergingStateFailed, err)
		return
	}

	currentNodeState, err := node.CurrentState()

	if err != nil {
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		return
	}

	err = Normalize(desiredState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		return
	}

	global.Machine.Transition(global.Running, nil)

	directoryStateUrl.Index = directoryState.Index // for next iteration
}

func Once(bootState *State) {
	Run(bootState, "5s")
}

func Loop(bootState *State, wait string) {
	for {
		Run(bootState, wait)
		time.Sleep(time.Second)
	}
}
