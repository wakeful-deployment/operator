package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/logger"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	var (
		nodeName   = flag.String("node", "", "The name of the host which is running operator")
		consulHost = flag.String("consul", "", "The name or ip of the consul host")
		configPath = flag.String("config", "./operator.json", "The path to the operator.json (default is .)")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "", "The timeout for polling")
		metadata   = flag.String("metadata", "", "JSON metadata to add to the directory for this node")
		verbose    = flag.Bool("verbose", false, "Log more info for easier debugging")
	)
	flag.Parse()

	state := LoadBootStateFromFile(*configPath)

	go runServer()

	// panic if config failed to load

	if global.Machine.IsCurrently(global.ConfigFailed) {
		panic(global.Machine.CurrentState.Error)
	}

	if *shouldLoop {
		state.ShouldLoop = true
	}

	// proceed with configuration

	if *nodeName != "" {
		state.NodeName = *nodeName
	}

	if *consulHost != "" {
		state.ConsulHost = *consulHost
	}

	// required flags

	if state.NodeName == "" || state.ConsulHost == "" {
		panic("ERROR: Must provide -node and -consul flags")
	}

	// other flags

	if *wait != "" {
		state.Wait = *wait
	}

	if *metadata != "" {
		var m map[string]string

		jsonErr := json.NewDecoder(strings.NewReader(*metadata)).Decode(&m)

		if jsonErr != nil {
			fmt.Println("-metadata was not valid json, skipping")
		} else {
			state.Metadata = m
		}
	}

	// defaults

	if state.Wait == "" {
		state.Wait = "5m"
	}

	logger.Verbose = *verbose

	// dependencies

	dockerClient := docker.EngineClient{}
	consulClient := consul.HttpClient{Host: state.ConsulHost}

	logger.Info("ready to go...")

	run(dockerClient, consulClient, state)
}

func runServer() {
	http.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, fmt.Sprintf("%v", global.Machine.CurrentState))
	})

	http.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		if global.Machine.IsCurrently(global.Running) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	http.ListenAndServe(":8000", nil)
}

func run(dockerClient docker.Client, consulClient consul.Client, state *State) {
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
