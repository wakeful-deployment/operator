package node

type Container struct {
	Name_    string
	Image_   string
	Ports_   []string
	Env_     map[string]string
	Restart_ string
	Tags_    []string
}

func (c Container) Name() string {
	return c.Name_
}

func (c Container) Image() string {
	return c.Image_
}

func (c Container) Ports() []string {
	return c.Ports_
}

func (c Container) Env() map[string]string {
	return c.Env_
}

func (c Container) Restart() string {
	return c.Restart_
}

func (c Container) Tags() []string {
	return c.Tags_
}
