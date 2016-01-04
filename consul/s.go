package consul

type S struct {
	Name_ string
	Tags_ []string
}

func (s S) Name() string {
	return s.Name_
}

func (s S) Tags() []string {
	return s.Tags_
}
