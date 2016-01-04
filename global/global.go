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
	Booting
	ConfigFailed
	Booted
	ConsulFailed
	NodeStateFailed
	MergeStateFailed
	NormalizeFailed
	DirectoryStateFailed
	Running
	Failing
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

var AllowedTransitions = fsm.Rules{
	fsm.From(Initial).To(Booting, ConfigFailed),
	fsm.From(Booting).To(Booted, ConsulFailed, NodeStateFailed, NormalizeFailed),
	fsm.From(ConsulFailed).To(Booting),
	fsm.From(NodeStateFailed).To(Booting, Running),
	fsm.From(MergeStateFailed).To(Booting, Running),
	fsm.From(NormalizeFailed).To(Booting, Running),
	fsm.From(DirectoryStateFailed).To(Booting, Running),
	fsm.From(Booted).To(ConsulFailed, NodeStateFailed, NormalizeFailed, Running, Failing),
	fsm.From(Running).To(Failing),
	fsm.From(Failing).To(Running, Booting),
}

func Transition(newState State) error {
	if newState.Const < Initial || newState.Const > Failing {
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
