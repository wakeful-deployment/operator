package boot

import (
	"errors"
	"fmt"
	// "github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/directory"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/global"
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

	fmt.Println(fmt.Sprintf("checking consul at %s", url))

	client := http.Client{Timeout: consulCheckTimeout}
	resp, err := client.Get(consulCheckUrl())

	if err == nil && resp.StatusCode == 200 {
		return nil
	}

	return errors.New("consul is not responding on port 8500")
}

func detectOrBootConsul(state *State) error {
	err := detectConsul()

	if err == nil {
		// consul was detected
		return nil
	}

	// it's not responding, so let's see if it's running

	containers, err := docker.RunningContainers()

	if err != nil {
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
		return errors.New("consul is running, but not responding on port 8500")
	}

	// it's not running, so let's try to boot it

	consulService := state.Services["consul"]

	if consulService.Name != "consul" {
		return errors.New("consul is not running. It is also not listed as a service, so cannot attempt to boot it")
	}

	consulContainer := consulService.Container()
	err = docker.Run(consulContainer)

	if err != nil {
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

func Boot(bootState *State) {
	if !global.Machine.IsCurrently(global.Booting) {
		global.Machine.Transition(global.Booting, nil)
	}

	err := detectOrBootConsul(bootState)

	if err != nil {
		global.Machine.Transition(global.ConsulFailed, err)
		return
	}

	// err = consul.PostMetadata(bootState.MetaData)
	// if err != nil {
	// 	global.Machine.Transition(global.PostingMetadataFailed, err)
	// 	return
	// }

	currentNodeState, err := node.CurrentState()

	if err != nil {
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		return
	}

	err = Normalize(bootState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		return
	}

	global.Machine.Transition(global.Booted, nil)
}

func GetState(wait string, index int) *directory.State {
	directoryStateUrl := directory.StateURL{Wait: wait}

	directoryState, err := directory.GetState(directoryStateUrl.String()) // this will block for some time

	if err != nil {
		global.Machine.Transition(global.FetchingDirectoryStateFailed, err)
		return nil
	}

	return directoryState
}

func Run(bootState *State, directoryState *directory.State) {
	if !global.Machine.IsCurrently(global.Running) && !global.Machine.IsCurrently(global.Booted) {
		global.Machine.Transition(global.AttemptingToRecover, global.Machine.CurrentState.Error)
	}

	desiredState, err := MergeStates(bootState, directoryState)

	if err != nil {
		global.Machine.Transition(global.MergingStateFailed, err)
		return
	}

	currentNodeState, err := node.CurrentState()

	if err != nil {
		global.Machine.Transition(global.FetchingNodeStateFailed, err)
		return
	}

	err = Normalize(desiredState, currentNodeState)

	if err != nil {
		global.Machine.Transition(global.NormalizingFailed, err)
		return
	}

	if !global.Machine.IsCurrently(global.Running) {
		global.Machine.Transition(global.Running, nil)
	}
}

func Once(bootState *State) {
	directoryState := GetState("0s", 0)
	Run(bootState, directoryState)
}

func Loop(bootState *State, wait string) {
	index := 0

	for {
		directoryState := GetState(wait, index)
		Run(bootState, directoryState)

		if global.Machine.IsCurrently(global.Running) {
			index = directoryState.Index
			time.Sleep(time.Second)
		} else {
			time.Sleep(6 * time.Second)
		}
	}
}
