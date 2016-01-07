package test

import (
	"github.com/wakeful-deployment/operator/container"
)

type DockerClient struct {
	RunResponse               error
	StopResponse              error
	RunningContainersFunction func() (string, error)
}

func (d DockerClient) Run(c container.Container) error {
	return d.RunResponse
}

func (d DockerClient) Stop(c container.Container) error {
	return d.StopResponse
}

func (d DockerClient) RunningContainers() (string, error) {
	result, err := d.RunningContainersFunction()

	if err != nil {
		return "", err
	}

	return result, nil
}
