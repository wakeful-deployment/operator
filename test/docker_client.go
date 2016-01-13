package test

import (
	"github.com/wakeful-deployment/operator/container"
)

type DockerClient struct {
	RunResponse               func(container.Container) error
	StopResponse              func(container.Container) error
	RunningContainersResponse func() (string, error)
}

func (d DockerClient) Run(c container.Container) error {
	return d.RunResponse(c)
}

func (d DockerClient) Stop(c container.Container) error {
	return d.StopResponse(c)
}

func (d DockerClient) RunningContainers() (string, error) {
	result, err := d.RunningContainersResponse()

	if err != nil {
		return "", err
	}

	return result, nil
}
