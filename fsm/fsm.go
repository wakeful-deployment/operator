package fsm

type State int

type Rule struct {
	From State
	To   []State
}

func (r Rule) Test(current State, next State) bool {
	if current != r.From {
		return false
	}

	for _, s := range r.To {
		if next == s {
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
