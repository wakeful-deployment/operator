package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type PortPair struct {
	Incoming int  `json:"incoming"`
	Outgoing int  `json:"outgoing"`
	UDP      bool `json:"udp"`
}
type Container struct {
	Name  string
	Image string
	Ports []PortPair
	Env   map[string]string
}

func (c Container) GetName() string {
	return c.Name
}

func (c Container) portsString() string {
	str := ""
	for _, portPair := range c.Ports {
		str = fmt.Sprintf("%s -p %d:%d", str, portPair.Incoming, portPair.Outgoing)
		if portPair.UDP {
			str = fmt.Sprintf("%s/udp", str)
		}
	}
	return str
}

func (c Container) envString() string {
	str := ""
	for key, value := range c.Env {
		str = fmt.Sprintf("%s -e %s=%s", str, key, value)
	}
	return str
}

func (c Container) Run() error {
	fmt.Printf("INFO: running container with name '%s' with image '%s'\n", c.Name, c.Image)

	args := []string{"run", "-d", "--name", c.Name}
	args = append(args, strings.Split(c.portsString(), " ")...)
	args = append(args, strings.Split(c.envString(), " ")...)
	args = append(args, c.Image)
	var cleaned []string
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			cleaned = append(cleaned, arg)
		}
	}

	_, err := exec.Command("docker", cleaned...).Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: 'docker run' failed: %v", err)
		return errors.New(errMsg)
	}

	return nil
}

func (c Container) Stop() error {
	fmt.Printf("INFO: stopping container with name '%s'\n", c.Name)

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

// Find keys that are present in first slice that are not present in second
func diffContainers(leftContainers []Container, rightContainers []Container) []Container {
	var diff []Container

	for _, left := range leftContainers {
		if isWhitelisted(left) {
			continue
		}

		// Let's assume at first it is missing
		isMissing := true

		for _, right := range rightContainers {
			if left.Name == right.Name {
				// If we find a match, then it's not missing
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			diff = append(diff, left)
		}
	}

	return diff
}

func RunningContainers() ([]Container, error) {
	psOut, err := exec.Command("docker", "ps", "--format", "{{.Names}} {{.Image}}").Output()

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: could not fetch running containers: %v\n", err)
		return nil, errors.New(errMsg)
	}

	return parseDockerPsOutput(string(psOut))
}

func parseDockerPsOutput(output string) ([]Container, error) {
	output = strings.TrimSpace(output)
	runningContainers := []Container{}

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

		name = info[0]  // First in the info
		image = info[1] // Second in the info
		if name == "operator" {
			continue
		}
		container := Container{Name: name, Image: image}
		runningContainers = append(runningContainers, container)
	}

	return runningContainers, nil
}

func NormalizeDockerContainers(desired []Container, current []Container) error {
	removed := diffContainers(current, desired)
	added := diffContainers(desired, current)

	fmt.Printf("INFO: Removed containers: %v\n", removed)
	fmt.Printf("INFO: Added containers: %v\n", added)

	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	errs := []error{}
	for _, container := range added {
		err := container.Run()
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, container := range removed {
		err := container.Stop()
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
