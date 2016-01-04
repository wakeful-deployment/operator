package main

import (
	"flag"
	"github.com/wakeful-deployment/operator/boot"
	"github.com/wakeful-deployment/operator/global"
	"time"
)

func main() {
	var (
		nodename   = flag.String("node", "", "The name of the host which is running operator")
		consulHost = flag.String("consul", "", "The name or ip of the consul host")
		configPath = flag.String("config", "./operator.json", "The path to the operator.json (default is .)")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "5m", "The timeout for polling")
	)
	flag.Parse()

	if *nodename == "" || *consulHost == "" {
		panic("ERROR: Must provide -node and -consul flags")
	}

	global.Info.Nodename = *nodename
	global.Info.Consulhost = *consulHost

	// TODO open a tcp port and respond with current state

	state := boot.LoadStateFromFile(*configPath)

	if global.CurrentState.Equal(global.ConfigFailed) {
		if *shouldLoop {
			for { // loop forever
			}
		} else {
			panic(global.CurrentState.Error)
		}
	}

	for {
		boot.Boot(state)

		if global.CurrentState.Equal(global.Booted) {
			break
		}

		time.Sleep(6 * time.Second)
	}

	if *shouldLoop {
		boot.Loop(state, *wait)
	} else {
		boot.Once(state)
	}
}
