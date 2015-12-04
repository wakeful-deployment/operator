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

type KeyState struct {
	Index int
	Keys  []ConsulKey
}

type ConsulKey struct {
	Key         string
	Value       string
	ModifyIndex int
}

func makeUrl(consulHost string, hostname string, index int) string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s?recurse=true&index=%d", consulHost, hostname, index)
}

func getKeyState(state *KeyState, url string) error {
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

		var keys []ConsulKey
		err = json.NewDecoder(resp.Body).Decode(&keys)

		if err != nil {
			return err
		}

		state.Keys = keys
		state.Index = index

		return nil
	case 404:
		index, err := strconv.Atoi(resp.Header["X-Consul-Index"][0])

		if err != nil {
			return err
		}

		state.Index = index
		state.Keys = []ConsulKey{}

		return nil
	default:
		return errors.New("non 200/404 response code")
	}
}

// Find keys that are present in first slice that are not present in second
func keyDiff(keySlice1 []ConsulKey, keySlice2 []ConsulKey) []ConsulKey {
	var diff []ConsulKey

	for _, keyInFirst := range keySlice1 {
		var present = true
		for _, keyInSecond := range keySlice2 {
			if keyInFirst.Key == keyInSecond.Key {
				present = false
				break
			}
		}

		if present {
			diff = append(diff, keyInFirst)
		}
	}

	return diff
}

func runningContainers() ([]string, error) {
	psOut, err := exec.Command("docker", "ps").Output()

	if err != nil {
		return []string{}, err
	}

	var runningContainers []string
	lines := strings.Split(string(psOut), "\n")
	for index, line := range lines {
		if index == 0 {
			continue
		}

		containerInfo := strings.Split(line, " ")
		containerName := containerInfo[len(containerInfo)-1]
		containerName = strings.Trim(containerName, " ")

		if len(containerName) > 0 {
			runningContainers = append(runningContainers, containerName)
		}
	}

	return runningContainers, nil
}

func containerIsRunning(containerName string, runningContainers []string) bool {
	containerIsRunning := false

	for _, runningContainerName := range runningContainers {
		if containerName == runningContainerName {
			containerIsRunning = true
		}
	}
	return containerIsRunning
}

func containerName(key ConsulKey) string {
	keyParts := strings.Split(key.Key, "/")
	return keyParts[len(keyParts)-1]
}

func imageName(key ConsulKey) string {
	base64Value := key.Value
	decoded, _ := base64.StdEncoding.DecodeString(base64Value)
	return string(decoded)
}

func runContainer(containerName string, imageName string) {
	fmt.Printf("Running container with name '%s' with image '%s'\n", containerName, imageName)
	_, err := exec.Command("docker", "run", "-d", "--name", containerName, imageName).Output()

	if err != nil {
		fmt.Println("ERROR: Could not run docker run successfully")
	}
}

func stopContainer(containerName string) {
	fmt.Printf("Stopping container with name '%s'\n", containerName)
	_, stopErr := exec.Command("docker", "stop", containerName).Output()

	if stopErr != nil {
		fmt.Println("ERROR: Could not run docker stop successfully")
	}

	time.Sleep(time.Second)

	_, rmErr := exec.Command("docker", "rm", containerName).Output()

	if rmErr != nil {
		fmt.Println("ERROR: Could not run docker rm successfully")
	}
}

func handleStateChange(previousState KeyState, newState KeyState) {
	// TODO: This find all keys in namespace that differ.
	// We want to only find 'app' keys
	removedKeys := keyDiff(previousState.Keys, newState.Keys)
	addedKeys := keyDiff(newState.Keys, previousState.Keys)

	fmt.Println("Removed:")
	fmt.Println(removedKeys)
	fmt.Println("Added:")
	fmt.Println(addedKeys)

	if len(addedKeys) == 0 && len(removedKeys) == 0 {
		return
	}

	runningContainers, err := runningContainers()

	if err != nil {
		//TODO should we handle this more gracefully
		fmt.Println(err)
		return
	}

	if len(addedKeys) > 0 {
		for _, key := range addedKeys {
			containerName := containerName(key)
			containerIsRunning := containerIsRunning(containerName, runningContainers)

			if !containerIsRunning {
				imageName := imageName(key)
				runContainer(containerName, imageName)
			}
		}
	}

	if len(removedKeys) > 0 {
		for _, key := range removedKeys {
			containerName := containerName(key)
			containerIsRunning := containerIsRunning(containerName, runningContainers)

			if containerIsRunning {
				stopContainer(containerName)
			}
		}
	}
}

func main() {
	hostname, hostnameExists := os.LookupEnv("HOSTNAME")
	consulHost, consulHostExists := os.LookupEnv("CONSUL_HOST")

	if !hostnameExists && !consulHostExists {
		panic("Must provide HOSTNAME and CONSUL_HOST env variables")
	}

	state := KeyState{Keys: []ConsulKey{}, Index: 0}

	for {
		newState := KeyState{}
		url := makeUrl(consulHost, hostname, state.Index)
		err := getKeyState(&newState, url)

		if err != nil {
			fmt.Println("There was an error getting the new state")
			syscall.Exit(1)
		}

		handleStateChange(state, newState)

		time.Sleep(time.Second)

		state = newState
	}
}
