package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Container struct {
	Name  string
	Image string
}

type ConsulState struct {
	Index int
	KVs   []ConsulKV
}

type ConsulKV struct {
	Key         string
	Value       string
	ModifyIndex int
}

func (c ConsulKV) ContainerName() string {
	keyParts := strings.Split(c.Key, "/")
	return keyParts[len(keyParts)-1]
}

func (c ConsulKV) ContainerImage() string {
	base64Value := c.Value
	decoded, _ := base64.StdEncoding.DecodeString(base64Value)
	return string(decoded)
}

func (c ConsulKV) ToContainer() Container {
	return Container{Name: c.ContainerName(), Image: c.ContainerImage()}
}

func (c ConsulState) Containers() []Container {
	var containers []Container

	for _, kv := range c.KVs {
		containers = append(containers, kv.ToContainer())
	}

	return containers
}

func makeUrl(consulHost string, hostname string, index int) string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s?recurse=true&index=%d", consulHost, hostname, index)
}

func getConsulStateBlocking(state *ConsulState, url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		index, err := strconv.Atoi(resp.Header["X-Consul-Index"][0])

		if err != nil {
			return err
		}

		var keys []ConsulKV
		err = json.NewDecoder(resp.Body).Decode(&keys)

		if err != nil {
			return err
		}

		state.KVs = keys
		state.Index = index

		return nil
	case 404:
		index, err := strconv.Atoi(resp.Header["X-Consul-Index"][0])

		if err != nil {
			return err
		}

		state.Index = index
		state.KVs = []ConsulKV{}

		return nil
	default:
		return errors.New("non 200/404 response code")
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

	for _, left := range leftContainers {
		if containerIsWhitelisted(left) {
			continue
		}

		for _, right := range rightContainers {
			if left.Name == right.Name {
				diff = append(diff, left)
				break
			}
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

func handleChange(newState ConsulState) {
	// TODO: This find all keys in namespace that differ.
	// We want to only find 'app' keys

	desired := newState.Containers()
	current, err := runningContainers()

	if err != nil {
		//TODO should we handle this more gracefully
		fmt.Printf("could not fetch running containers: %v\n", err)
		return
	}

	removed := diffContainers(current, desired)
	added := diffContainers(desired, current)

	fmt.Printf("Removed: %v\n", removed)
	fmt.Printf("Added: %v\n", added)

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

func main() {
	hostname, hostnameExists := os.LookupEnv("HOSTNAME")
	consulHost, consulHostExists := os.LookupEnv("CONSUL_HOST")

	if !hostnameExists && !consulHostExists {
		panic("Must provide HOSTNAME and CONSUL_HOST env variables")
	}

	lastIndex := 0

	for {
		desiredState := ConsulState{}
		url := makeUrl(consulHost, hostname, lastIndex)
		err := getConsulStateBlocking(&desiredState, url)

		if err != nil {
			fmt.Println("There was an error getting the new state")
			syscall.Exit(1)
		}

		handleChange(desiredState)

		time.Sleep(time.Second)

		lastIndex = desiredState.Index
	}
}
