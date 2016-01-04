package docker

type C struct {
	Name_    string
	Image_   string
	Ports_   []string
	Env_     map[string]string
	Restart_ string
	Tags_    []string
}

func (c C) Name() string {
	return c.Name_
}

func (c C) Image() string {
	return c.Image_
}

func (c C) Ports() []string {
	return c.Ports_
}

func (c C) Env() map[string]string {
	return c.Env_
}

func (c C) Restart() string {
	return c.Restart_
}

func (c C) Tags() []string {
	return c.Tags_
}
