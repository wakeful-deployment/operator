package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wakeful-deployment/operator/boot"
	"github.com/wakeful-deployment/operator/global"
	"strings"
	"time"
)

func main() {
	var (
		nodename   = flag.String("node", "", "The name of the host which is running operator")
		consulHost = flag.String("consul", "", "The name or ip of the consul host")
		configPath = flag.String("config", "./operator.json", "The path to the operator.json (default is .)")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "", "The timeout for polling")
		metadata   = flag.String("metadata", "", "JSON metadata to add to the directory for this node")
	)
	flag.Parse()

	// TODO open a tcp port and respond with current state

	state := boot.LoadBootStateFromFile(*configPath)

	if *shouldLoop {
		state.ShouldLoop = true
	}

	// panic or loop if config failed to load

	if global.Machine.IsCurrently(global.ConfigFailed) {
		if state.ShouldLoop {
			for { // loop forever
			}
		} else {
			panic(global.Machine.CurrentState.Error)
		}
	}

	// proceed with configuration

	if *nodename != "" {
		state.Nodename = *nodename
	}

	if *consulHost != "" {
		state.ConsulHost = *consulHost
	}

	// required flags

	if state.Nodename == "" || state.ConsulHost == "" {
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
		}

		state.MetaData = m
	}

	// defaults

	if state.Wait == "" {
		state.Wait = "5m"
	}

	// globals

	global.Info.Nodename = state.Nodename
	global.Info.Consulhost = state.ConsulHost

	// boot

	for {
		boot.Boot(state)

		if global.Machine.IsCurrently(global.Booted) {
			break
		}

		time.Sleep(6 * time.Second)
	}

	// run

	if *shouldLoop {
		boot.Loop(state, *wait)
	} else {
		boot.Once(state)
	}
}
