package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Container struct {
	Name  string
	Image string
}

func (c Container) IsPresent(containers []Container) bool {
	for _, other := range containers {
		if other.Name == c.Name {
			return true
		}
	}

	return false
}

func (c Container) Run() {
	fmt.Printf("Running container with name '%s' with image '%s'\n", c.Name, c.Image)
	_, err := exec.Command("docker", "run", "-d", "-P", "--name", c.Name, c.Image).Output()

	if err != nil {
		fmt.Println("ERROR: Could not run docker run successfully")
	}
}

func (c Container) Stop() {
	fmt.Printf("Stopping container with name '%s'\n", c.Name)
	_, stopErr := exec.Command("docker", "stop", c.Name).Output()

	if stopErr != nil {
		fmt.Println("ERROR: Could not run docker stop successfully")
	}

	time.Sleep(time.Second)

	_, rmErr := exec.Command("docker", "rm", c.Name).Output()

	if rmErr != nil {
		fmt.Println("ERROR: Could not run docker rm successfully")
	}
}

func containerIsWhitelisted(container Container) bool {
	containerWhiteList := []string{"consul", "statsite", "operator"}
	for _, name := range containerWhiteList {
		if container.Name == name {
			return true
		}
	}

	return false
}

// Find keys that are present in first slice that are not present in second
func diffContainers(leftContainers []Container, rightContainers []Container) []Container {
	var diff []Container

	fmt.Println(fmt.Sprintf("compared %v to %v", leftContainers, rightContainers))

	for _, left := range leftContainers {
		if containerIsWhitelisted(left) {
			continue
		}

		// Let's assume at first it is missing
		isMissing := true

		for _, right := range rightContainers {
			fmt.Println(fmt.Sprintf("comparing %s to %s", left.Name, right.Name))
			if left.Name == right.Name {
				// If we find a match, then it's not missing, but is present
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			fmt.Println(fmt.Sprintf("adding %s to the diff", left.Name))
			diff = append(diff, left)
		}
	}

	return diff
}

func runningContainers() ([]Container, error) {
	psOut, err := exec.Command("docker", "ps").Output()

	if err != nil {
		return []Container{}, err
	}

	var runningContainers []Container

	lines := strings.Split(string(psOut), "\n")

	for index, line := range lines {
		if index == 0 {
			continue
		}

		info := strings.Split(line, " ")

		var name string
		var image string

		if len(info) < 2 {
			fmt.Printf("retreived info was not formatted correctly: %v\n", info)
			continue
		} else {
			fmt.Printf("info: %v\n", info)

			name = info[len(info)-1]
			name = strings.Trim(name, " ")

			image = info[1]
			image = strings.Trim(image, " ")
		}

		if len(name) > 0 {
			container := Container{Name: name, Image: image}
			runningContainers = append(runningContainers, container)
		}
	}

	return runningContainers, nil
}

func normalizeDockerContainers(newState ConsulState) {
	// TODO: This find all keys in namespace that differ.
	// We want to only find 'app' keys

	desired := newState.Containers()
	current, err := runningContainers()

	fmt.Println(fmt.Sprintf("current running containers: %v", current))

	if err != nil {
		//TODO should we handle this more gracefully
		fmt.Printf("could not fetch running containers: %v\n", err)
		return
	}

	removed := diffContainers(current, desired)
	added := diffContainers(desired, current)

	fmt.Printf("Removed containers: %v\n", removed)
	fmt.Printf("Added containers: %v\n", added)

	if len(added) == 0 && len(removed) == 0 {
		return
	}

	if len(added) > 0 {
		for _, container := range added {
			container.Run()
		}
	}

	if len(removed) > 0 {
		for _, container := range removed {
			container.Stop()
		}
	}
}
