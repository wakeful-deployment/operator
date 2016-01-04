package docker

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func Diff(left []Container, right []Container) []Container {
	var result []Container

	for _, leftItem := range left {
		// Let's assume at first it is missing
		isMissing := true

		for _, rightItem := range right {
			if leftItem.Name() == rightItem.Name() {
				// If we find a match, then it's not missing
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			result = append(result, leftItem)
		}
	}

	return result
}

func Run(c Container) error {
	fmt.Printf("INFO: running container with name '%s' with image '%s'\n", c.Name(), c.Image())

	args := RunArgs(c)
	_, err := exec.Command("docker", args...).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker run' failed: %v", err)
		return errors.New(errMsg)
	}

	return nil
}

func Stop(c Container) error {
	fmt.Printf("INFO: stopping container with name '%s'\n", c.Name)

	_, err := exec.Command("docker", "stop", c.Name()).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker stop' failed: %v", err)
		return errors.New(errMsg)
	}

	time.Sleep(time.Second)

	_, err = exec.Command("docker", "rm", c.Name()).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker rm' failed: %v", err)
		return errors.New(errMsg)
	}

	return nil
}

// For now, we assume if it's running then it's running with the correct args. It's possible in the future we will inspect each container and compare every arg.

func RunningContainers() ([]C, error) {
	psOut, err := exec.Command("docker", "ps", "--format", "{{.Names}} {{.Image}}").Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: could not fetch running containers: %v\n", err)
		return nil, errors.New(errMsg)
	}

	return parseDockerPsOutput(string(psOut))
}

func parseDockerPsOutput(output string) ([]C, error) {
	output = strings.TrimSpace(output)
	runningContainers := []C{}

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

		container := C{Name_: name, Image_: image}
		runningContainers = append(runningContainers, container)
	}

	return runningContainers, nil
}

func NormalizeDockerContainers(desired []Container, current []Container) error {
	removed := Diff(current, desired)
	added := Diff(desired, current)

	fmt.Printf("INFO: Removed containers: %v\n", removed)
	fmt.Printf("INFO: Added containers: %v\n", added)

	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	errs := []error{}
	for _, container := range added {
		err := Run(container)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, container := range removed {
		err := Stop(container)
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
