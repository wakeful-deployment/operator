package node

type ConsulService struct {
	Name_ string
	Tags_ []string
}

func (c ConsulService) Name() string {
	return c.Name_
}

func (c ConsulService) Tags() []string {
	return c.Tags_
}
