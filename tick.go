package main

import (
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

func Once(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	directoryState := GetDirectoryState(consulClient, bootState.NodeName, 0, "0s")
	Tick(dockerClient, consulClient, bootState, directoryState)
}

func Loop(dockerClient docker.Client, consulClient consul.Client, bootState *State) {
	index := 0

	for {
		directoryState := GetDirectoryState(consulClient, bootState.NodeName, index, bootState.Wait)
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

func GetDirectoryState(consulClient consul.Client, nodeName string, index int, wait string) *consul.DirectoryState {
	logger.Info("getting directory state...")
	directoryState, err := consulClient.GetDirectoryState(nodeName, index, wait) // this will block for some time

	if err != nil {
		logger.Error(fmt.Sprintf("fetching directory state failed with error: %v", err))
		global.Machine.Transition(global.FetchingDirectoryStateFailed, err)
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
		logger.Error(fmt.Sprintf("merging states failed with error: %v", err))
		global.Machine.Transition(global.MergingStateFailed, err)
		return
	}

	logger.Info("getting current node state")
	currentNodeState, err := node.CurrentState(dockerClient, consulClient)

	if err != nil {
		logger.Error(fmt.Sprintf("getting current node state failed with error: %v", err))
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		return
	}

	logger.Info(fmt.Sprintf("normalizing states - desiredState=%v and currentNodeState=%v", desiredState, currentNodeState))
	err = normalize(dockerClient, consulClient, desiredState, currentNodeState)

	if err != nil {
		logger.Error(fmt.Sprintf("normalizing failed with error: %v", err))
		global.Machine.Transition(global.NormalizingFailed, err)
		return
	}

	if !global.Machine.IsCurrently(global.Running) {
		global.Machine.Transition(global.Running, nil)
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
