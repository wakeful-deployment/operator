package docker

type ContainerCollectionI interface {
	Len() int
	At(index int) ContainerI
	Append(ContainerI)
}

type ContainerCollection []Container

func (c ContainerCollection) Len() int {
	return len(c)
}

func (c ContainerCollection) At(index int) Container {
	return c[index]
}

func (c ContainerCollection) Append(container Container) {
	c = append(c, container)
}

type ContainerI interface {
	Name() string
	Image() string
	Ports() []string
	Env() map[string]string
	Restart() string
	Tags() []string
}

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
