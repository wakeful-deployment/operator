package boot

import (
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/directory"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/logger"
	"github.com/wakeful-deployment/operator/node"
	"net/http"
	"time"
)

func consulCheckUrl() string {
	return fmt.Sprintf("http://%s:8500/", global.Info.Consulhost)
}

const consulCheckTimeout = 5 * time.Second

func detectConsul() error {
	url := consulCheckUrl()

	logger.Info(fmt.Sprintf("checking consul with url = %s", url))

	client := http.Client{Timeout: consulCheckTimeout}
	resp, err := client.Get(consulCheckUrl())

	if err == nil && resp.StatusCode == 200 {
		return nil
	}

	if err != nil {
		logger.Error(fmt.Sprintf("checking consul failed with error: %v", err))
	} else {
		logger.Error(fmt.Sprintf("checking consul failed with non-200 response: %d", resp.StatusCode))
	}

	return errors.New("consul is not responding on port 8500")
}

func detectOrBootConsul(dockerClient docker.Client, state *State) error {
	err := detectConsul()

	if err == nil {
		logger.Info("consul detected!")
		return nil
	}

	logger.Error(fmt.Sprintf("detecting consul failed with error: %v. Let's check docker to see if it's running", err))

	containers, err := docker.RunningContainers(dockerClient)

	if err != nil {
		logger.Error(fmt.Sprintf("detecting docker failed also with error: %v.", err))
		return errors.New("consul and docker are both not responding")
	}

	running := false
	for _, container := range containers {
		if container.Name == "consul" {
			running = true
			break
		}
	}

	if running {
		logger.Error("consul is running, but we already detected it is not responding on port 8500")
		return errors.New("consul is running, but not responding on port 8500")
	}

	logger.Info("consul not running. Attempting now to boot it up")

	consulService := state.Services["consul"]

	if consulService.Name != "consul" {
		logger.Error("could not even find consul as a service when trying to boot it up.")
		return errors.New("consul is not running. It is also not listed as a service, so cannot attempt to boot it")
	}

	consulContainer := consulService.Container(state.NodeName)
	err = dockerClient.Run(consulContainer)

	if err != nil {
		logger.Error(fmt.Sprintf("attemping to run consul with docker failed with error: %v", err))
		return err
	}

	time.Sleep(time.Second)                // give docker some time to make sure it would show up in the process list
	return errors.New("consul is booting") // the caller of this function will attempt again which should attempt to detect consul
}

func LoadBootStateFromFile(path string) *State {
	state, err := ReadStateFromConfigFile(path)

	if err != nil {
		global.Machine.Transition(global.ConfigFailed, err)
		return nil
	}

	return state
}

func Boot(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	if !global.Machine.IsCurrently(global.Booting) {
		logger.Info("booting up...")
		global.Machine.Transition(global.Booting, nil)
	}

	logger.Info("checking status of consul...")
	err := detectOrBootConsul(dockerClient, bootState)

	if err != nil {
		global.Machine.Transition(global.ConsulFailed, err)
		logger.Error(fmt.Sprintf("detecting or booting consul failed with error: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("posting metadata to consul. Metadata = %v", bootState.Metadata))
	err = consulClient.PostMetadata(bootState.NodeName, bootState.Metadata)

	if err != nil {
		global.Machine.Transition(global.PostingMetadataFailed, err)
		logger.Error(fmt.Sprintf("posting metadata failed with error: %v", err))
		return
	}

	logger.Info("getting current node state")
	currentNodeState, err := node.CurrentState(dockerClient, consulClient)

	if err != nil {
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		logger.Error(fmt.Sprintf("fetching current node state failed with error: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("normalizing states - bootState=%v and currentNodeState=%v", bootState, currentNodeState))
	err = Normalize(dockerClient, consulClient, bootState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		logger.Error(fmt.Sprintf("normalizing states failed with error: %v", err))
		return
	}

	global.Machine.Transition(global.Booted, nil)
	logger.Info("booted!")
}

func GetState(nodeName string, wait string, index int) *directory.State {
	directoryStateUrl := directory.StateURL{Wait: wait, Index: index}

	logger.Info(fmt.Sprintf("getting directory state with url: %s", directoryStateUrl.String(nodeName)))
	directoryState, err := directory.GetState(directoryStateUrl.String(nodeName)) // this will block for some time

	if err != nil {
		global.Machine.Transition(global.FetchingDirectoryStateFailed, err)
		logger.Error(fmt.Sprintf("fetching directory state failed with error: %v", err))
		return nil
	}
	logger.Info(fmt.Sprintf("succesfully fetched directoryState: %v", directoryState))

	return directoryState
}

func Run(dockerClient docker.Client, consulClient consul.Client, bootState *State, directoryState *directory.State) {
	if !global.Machine.IsCurrently(global.Running) && !global.Machine.IsCurrently(global.Booted) {
		global.Machine.Transition(global.AttemptingToRecover, global.Machine.CurrentState.Error)
	}

	logger.Info(fmt.Sprintf("merging states - bootState=%v and directoryState=%v", bootState, directoryState))
	desiredState, err := MergeStates(bootState, directoryState)

	if err != nil {
		global.Machine.Transition(global.MergingStateFailed, err)
		logger.Error(fmt.Sprintf("merging states failed with error: %v", err))
		return
	}

	logger.Info("getting current node state")
	currentNodeState, err := node.CurrentState(dockerClient, consulClient)

	if err != nil {
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		logger.Error(fmt.Sprintf("getting current node state failed with error: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("normalizing states - desiredState=%v and currentNodeState=%v", desiredState, currentNodeState))
	err = Normalize(dockerClient, consulClient, desiredState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		logger.Error(fmt.Sprintf("normalizing failed with error: %v", err))
		return
	}

	if !global.Machine.IsCurrently(global.Running) {
		global.Machine.Transition(global.Running, nil)
	}
}

func Once(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	directoryState := GetState(bootState.NodeName, "0s", 0)
	Run(dockerClient, consulClient, bootState, directoryState)
}

func Loop(dockerClient docker.Client, consulClient consul.Client, bootState *State, wait string) {
	index := 0

	for {
		directoryState := GetState(bootState.NodeName, wait, index)
		Run(dockerClient, consulClient, bootState, directoryState)

		if global.Machine.IsCurrently(global.Running) {
			logger.Info(fmt.Sprintf("iteration complete - setting index to %d and then sleeping", directoryState.Index))
			index = directoryState.Index
			time.Sleep(time.Second)
		} else {
			logger.Info(fmt.Sprintf("iteration complete - machine is not running state but rather %v. Sleeping now.", global.Machine.CurrentState))
			time.Sleep(6 * time.Second)
		}
	}
}
