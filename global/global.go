package global

import (
	"fmt"
	"github.com/wakeful-deployment/operator/fsm"
)

type GlobalInfo struct {
	Nodename   string
	Consulhost string
}

type GlobalMetadata map[string]string

const (
	Initial fsm.State = 0 + iota
	ConfigFailed
	Booting
	ReattemptingToBoot
	Booted
	ConsulFailed
	FetchingNodeStateFailed
	MergingStateFailed
	NormalizingFailed
	FetchingDirectoryStateFailed
	AttemptingToRecover
	Running
)

type State struct {
	Const fsm.State
	Error error
}

func (s State) Equal(c fsm.State) bool {
	return s.Const == c
}

func (s State) NotEqual(c fsm.State) bool {
	return s.Const != c
}

func NewState(s fsm.State, e error) State {
	return State{Const: s, Error: e}
}

var CurrentState = State{Const: Initial}

var Failures = []fsm.State{ConsulFailed, FetchingNodeStateFailed, NormalizingFailed}

var AllowedTransitions = fsm.Rules{
	fsm.From(Initial).To(Booting, ConfigFailed),
	fsm.From(Booting).To(append(Failures, Booted)...),
	fsm.From(ReattemptingToBoot).To(append(Failures, Booted)...),
	fsm.From(ConsulFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(FetchingNodeStateFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(MergingStateFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(NormalizingFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(FetchingDirectoryStateFailed).To(ReattemptingToBoot, AttemptingToRecover),
	fsm.From(Booted).To(append(Failures, Running)...),
	fsm.From(Running).To(append(Failures, Running)...),
}

func Transition(newState State) error {
	if newState.Const < Initial || newState.Const > Running {
		panic("FATAL ERROR: Cannot transition to illegal state")
	}

	if AllowedTransitions.Test(CurrentState.Const, newState.Const) {
		fmt.Println(fmt.Sprintf("Transitioned from %v to %v", CurrentState, newState))
		CurrentState = newState
		return nil
	} else {
		panic(fmt.Sprintf("FATAL ERROR: Cannot transition from %s to %s, not allowed", newState, CurrentState))
	}
}

var Info = GlobalInfo{}
var Metadata = GlobalMetadata{}
