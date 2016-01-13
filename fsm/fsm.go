package fsm

import (
	"fmt"
)

type Machine struct {
	CurrentState State
	Rules        Rules
	States       []State
}

func (m *Machine) ForceTransition(to State, e error) {
	to.Error = e
	fmt.Println(fmt.Sprintf("FSM: force transitioned from %v to %v", m.CurrentState, to))
	m.CurrentState = to
}

func (m *Machine) Transition(to State, e error) {
	stateIsPresent := false
	for _, s := range m.States {
		if to.Equal(s) {
			stateIsPresent = true
		}
	}

	if !stateIsPresent {
		panic(fmt.Sprintf("FATAL ERROR: Cannot transition to illegal state: %v", to))
	}

	if m.Rules.Test(m.CurrentState, to) {
		to.Error = e
		fmt.Println(fmt.Sprintf("FSM: transitioned from %v to %v", m.CurrentState, to))
		m.CurrentState = to
	} else {
		panic(fmt.Sprintf("FATAL ERROR: Cannot transition from %s to %s, not allowed", m.CurrentState, to))
	}
}

func (m Machine) IsCurrently(s State) bool {
	return m.CurrentState.Equal(s)
}

type State struct {
	Name  string
	Error error
}

func (s State) Equal(other State) bool {
	return s.Name == other.Name
}

func (s State) NotEqual(other State) bool {
	return s.Name != other.Name
}

type Rule struct {
	From State
	To   []State
}

func (r Rule) Test(current State, next State) bool {
	if current.NotEqual(r.From) {
		return false
	}

	for _, s := range r.To {
		if next.Equal(s) {
			return true
		}
	}

	return false
}

type PartialRule struct {
	From State
}

func (p PartialRule) To(to ...State) Rule {
	return Rule{From: p.From, To: to}
}

func From(s State) PartialRule {
	return PartialRule{From: s}
}

type Rules []Rule

func (r Rules) Test(current State, next State) bool {
	for _, rule := range r {
		if rule.Test(current, next) {
			return true
		}
	}

	return false
}
