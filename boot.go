package main

import (
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/logger"
	"time"
)

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

	global.Machine.Transition(global.Booted, nil)
	logger.Info("booted!")
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
