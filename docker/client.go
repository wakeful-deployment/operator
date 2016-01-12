package docker

import (
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/logger"
	"os/exec"
	"time"
)

type Client interface {
	Run(container.Container) error
	Stop(container.Container) error
	RunningContainers() (string, error)
}

type EngineClient struct{}

func (d EngineClient) Run(c container.Container) error {
	logger.Info(fmt.Sprintf("running container with name '%s' with image '%s'", c.Name, c.Image))

	args := RunArgs(c)
	_, err := exec.Command("docker", args...).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker run' failed: %v", err)
		return errors.New(errMsg)
	}

	return nil
}

func (d EngineClient) Stop(c container.Container) error {
	logger.Info(fmt.Sprintf("stopping container with name '%s'", c.Name))

	_, err := exec.Command("docker", "stop", c.Name).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker stop' failed: %v", err)
		return errors.New(errMsg)
	}

	time.Sleep(time.Second)

	_, err = exec.Command("docker", "rm", c.Name).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker rm' failed: %v", err)
		return errors.New(errMsg)
	}

	return nil
}

// For now, we assume if it's running then it's running with the correct args. It's possible in the future we will inspect each container and compare every arg.

func (d EngineClient) RunningContainers() (string, error) {
	psOut, err := exec.Command("docker", "ps", "--format", "{{.Names}} {{.Image}}").Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: could not fetch running containers: %v\n", err)
		return "", errors.New(errMsg)
	}

	return string(psOut), nil
}
