package main

import (
	"flag"
	"fmt"
	"time"
)

func tick(stateUrl *ConsulStateURL, servicesUrl *ConsulServicesURL) (int, error) {
	desiredState, err := NewConsulState(stateUrl.ToString())

	if err != nil {
		fmt.Sprintf("ERROR: error getting the new consul KV state: %v", err)
		return desiredState.Index, err
	}

	services, sErr := GetConsulServices(*servicesUrl)

	if sErr != nil {
		fmt.Sprintf("ERROR: error getting the registered consul services: %v", err)
		return desiredState.Index, sErr
	}

	normalizeDockerContainers(*desiredState)
	normalizeConsulServices(*desiredState, services, stateUrl.Host)

	return desiredState.Index, nil
}

func loop(stateUrl *ConsulStateURL, servicesUrl *ConsulServicesURL) {
	for {
		newIndex, err := tick(stateUrl, servicesUrl)
		stateUrl.Index = newIndex

		time.Sleep(time.Second)

		if err != nil {
			continue
		}
	}
}

func main() {
	var (
		hostname   = flag.String("hostname", "", "The name of the host which is running operator")
		consulHost = flag.String("consulhost", "", "The name of the consul host")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "5m", "The timeout for polling")
	)
	flag.Parse()

	if *hostname == "" && *consulHost == "" {
		panic("ERROR: Must provide hostname and consulhost flags")
	}

	stateUrl := ConsulStateURL{Host: *consulHost, Hostname: *hostname, Wait: *wait}
	servicesUrl := ConsulServicesURL{Host: *consulHost}

	if *shouldLoop {
		loop(&stateUrl, &servicesUrl)
	} else {
		tick(&stateUrl, &servicesUrl)
	}
}
