package fsm

import (
	"testing"
)

var state1 State = State{Name: "state1"}
var state2 State = State{Name: "state2"}
var state3 State = State{Name: "state3"}
var states []State = []State{
	state1,
	state2,
	state3,
}

var rules Rules = Rules{
	From(state1).To(state2),
	From(state2).To(state3),
	From(state3).To(state1, state2),
}

func machine() Machine {
	return Machine{CurrentState: state1, Rules: rules, States: states}
}

func TestMultipleTransitions(t *testing.T) {
	m := machine()

	m.Transition(state2, nil)

	if !m.IsCurrently(state2) {
		t.Errorf("State is expected to be State 2 but was %s", m.CurrentState.Name)
	}

	m.Transition(state3, nil)

	if !m.IsCurrently(state3) {
		t.Errorf("State is expected to be State 3 but was %s", m.CurrentState.Name)
	}

	m.Transition(state1, nil)

	if !m.IsCurrently(state1) {
		t.Errorf("State is expected to be State 1 but was %s", m.CurrentState.Name)
	}
}

func TestIllegalTransition(t *testing.T) {
	// Catch panics
	var err interface{}
	defer func() {
		err = recover()
	}()
	m := machine()

	m.Transition(state2, nil)
	m.Transition(state1, nil)

	if err == nil {
		t.Errorf("expected illegal transition to panic but did not")
	}
}
