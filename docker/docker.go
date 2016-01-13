package docker

import (
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/logger"
	"strings"
)

func RunningContainers(client Client) ([]container.Container, error) {
	output, err := client.RunningContainers()

	if err != nil {
		return nil, err
	}

	return parseDockerPsOutput(output)
}

func NormalizeContainers(client Client, desired []container.Container, current []container.Container) error {
	removed := container.Diff(current, desired)
	added := container.Diff(desired, current)

	logger.Info(fmt.Sprintf("removed containers: %v", removed))
	logger.Info(fmt.Sprintf("added containers: %v", added))

	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	errs := []error{}
	for _, container := range added {
		err := client.Run(container)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, container := range removed {
		err := client.Stop(container)

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		errMsg := fmt.Sprintf("ERROR: At least 1 error normalizing containers: %v", errs)
		return errors.New(errMsg)
	}

	return nil
}

func parseDockerPsOutput(output string) ([]container.Container, error) {
	output = strings.TrimSpace(output)
	var runningContainers []container.Container

	if output == "" {
		return runningContainers, nil
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		info := strings.Split(line, " ")
		var cleanedInfo []string

		for _, str := range info {
			str = strings.TrimSpace(str)
			if str != "" {
				cleanedInfo = append(cleanedInfo, str)
			}
		}
		info = cleanedInfo

		var name string
		var image string

		if len(info) != 2 {
			errMsg := fmt.Sprintf("ERROR: 'docker ps' info was not formatted correctly: %s\n", line)
			return nil, errors.New(errMsg)
		}

		name = info[0]
		image = info[1]

		if name == "operator" {
			continue
		}

		container := container.Container{Name: name, Image: image}
		runningContainers = append(runningContainers, container)
	}

	return runningContainers, nil
}
