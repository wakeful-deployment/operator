package main

import (
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/logger"
	"github.com/wakeful-deployment/operator/node"
	"github.com/wakeful-deployment/operator/service"
	"time"
)

func Run(dockerClient docker.Client, consulClient consul.Client, state *State) {
	for {
		Boot(dockerClient, consulClient, state)

		if global.Machine.IsCurrently(global.Booted) {
			break
		}

		time.Sleep(6 * time.Second)
	}

	if state.ShouldLoop {
		Loop(dockerClient, consulClient, state)
	} else {
		Once(dockerClient, consulClient, state)
	}
}

func LoadBootStateFromFile(path string) *State {
	state, err := ReadStateFromConfigFile(path)

	if err != nil {
		global.Machine.Transition(global.ConfigFailed, err)
		return nil
	}

	for name, s := range state.Services {
		s.Name = name
	}

	return state
}

func Boot(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	if !global.Machine.IsCurrently(global.Booting) {
		logger.Info("booting up...")
		global.Machine.Transition(global.Booting, nil)
	}

	logger.Info("checking status of consul...")
	err := detectOrBootConsul(dockerClient, consulClient, bootState)

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
	err = normalize(dockerClient, consulClient, bootState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		logger.Error(fmt.Sprintf("normalizing states failed with error: %v", err))
		return
	}

	global.Machine.Transition(global.Booted, nil)
	logger.Info("booted!")
}

func Once(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	directoryState := GetState(consulClient, bootState.NodeName, 0, "0s")
	Tick(dockerClient, consulClient, bootState, directoryState)
}

func Loop(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	index := 0

	for {
		directoryState := GetState(consulClient, bootState.NodeName, index, bootState.Wait)
		Tick(dockerClient, consulClient, bootState, directoryState)

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

// reconcile the desired config with the current state
func normalize(dockerClient docker.Client, consulClient consul.Client, desiredState *State, currentNodeState *node.State) error {
	// always try to fix the containers before fixing the registrations

	var desiredContainers []container.Container

	for _, service := range desiredState.Services {
		desiredContainers = append(desiredContainers, service.Container(desiredState.NodeName, consulClient.ConsulHost()))
	}

	err := docker.NormalizeContainers(dockerClient, desiredContainers, currentNodeState.Containers)

	if err != nil {
		return err
	}

	// then try to register everything correctly in consul

	var desiredServices []service.Service

	for _, s := range desiredState.Services {
		desiredServices = append(desiredServices, *s)
	}

	err = consul.NormalizeServices(consulClient, desiredServices, currentNodeState.Services)

	if err != nil {
		return err
	}

	return nil
}

func GetState(consulClient consul.Client, nodeName string, index int, wait string) *consul.DirectoryState {
	logger.Info("getting directory state...")
	directoryState, err := consulClient.GetDirectoryState(nodeName, index, wait) // this will block for some time

	if err != nil {
		global.Machine.Transition(global.FetchingDirectoryStateFailed, err)
		logger.Error(fmt.Sprintf("fetching directory state failed with error: %v", err))
		return nil
	}
	logger.Info(fmt.Sprintf("succesfully fetched directoryState: %v", directoryState))

	return directoryState
}

func Tick(dockerClient docker.Client, consulClient consul.Client, bootState *State, directoryState *consul.DirectoryState) {
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
	err = normalize(dockerClient, consulClient, desiredState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		logger.Error(fmt.Sprintf("normalizing failed with error: %v", err))
		return
	}

	if !global.Machine.IsCurrently(global.Running) {
		global.Machine.Transition(global.Running, nil)
	}
}

func detectOrBootConsul(dockerClient docker.Client, consulClient consul.Client, state *State) error {
	logger.Info("checking consul...")

	err := consulClient.Detect()

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

	if consulService == nil {
		logger.Error("need to boot consul, but consul is not one of the services specified in the desired state")
		return errors.New("need to boot consul, but consul is not one of the services specified in the desired state")
	}

	if consulService.Name != "consul" {
		logger.Error("could not even find consul as a service when trying to boot it up.")
		return errors.New("consul is not running. It is also not listed as a service, so cannot attempt to boot it")
	}

	consulContainer := consulService.Container(state.NodeName, consulClient.ConsulHost())
	err = dockerClient.Run(consulContainer)

	if err != nil {
		logger.Error(fmt.Sprintf("attemping to run consul with docker failed with error: %v", err))
		return err
	}

	time.Sleep(time.Second)                // give docker some time to make sure it would show up in the process list
	return errors.New("consul is booting") // the caller of this function will attempt again which should attempt to detect consul
}
