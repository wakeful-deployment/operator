package main

import (
	"flag"
	"fmt"
	"time"
)

type ConsulStateURL struct {
	Host     string
	Hostname string
	Index    int
	Wait     string
}

type ConsulServicesURL struct {
	Host string
}

func (c *ConsulStateURL) ToString() string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s?recurse=true&index=%d&wait=%s", c.Host, c.Hostname, c.Index, c.Wait)
}

func (c *ConsulServicesURL) ToString() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/services", c.Host)
}

func tick(stateUrl *ConsulStateURL, servicesUrl *ConsulServicesURL) (int, error) {
	fmt.Println("tick!")

	desiredState := ConsulState{}
	err := GetConsulState(&desiredState, stateUrl.ToString())

	fmt.Println(fmt.Sprintf("desiredState: %v", desiredState))

	if err != nil {
		fmt.Sprintf("There was an error getting the new consul state: %v", err)
		return desiredState.Index, err
	}

	services, sErr := GetConsulServices(*servicesUrl)

	if sErr != nil {
		fmt.Sprintf("There was an error getting the registered consul services: %v", err)
		return desiredState.Index, sErr
	}

	normalizeDockerContainers(desiredState)
	normalizeConsulServices(desiredState, services)

	fmt.Println("tock!")

	return desiredState.Index, nil
}

func loop(stateUrl *ConsulStateURL, servicesUrl *ConsulServicesURL) {
	for {
		fmt.Println("begin the for")
		newIndex, err := tick(stateUrl, servicesUrl)
		stateUrl.Index = newIndex

		time.Sleep(time.Second)

		if err != nil {
			fmt.Println(fmt.Sprintf("error!!!! %v", err))
			continue
		}
		fmt.Println("end the for")
	}
}

func main() {
	var hostname = flag.String("hostname", "", "The name of the host which is running operator")
	var consulHost = flag.String("consulhost", "", "The name of the consul host")
	var wait = flag.String("wait", "5m", "The timeout for polling")
	var shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
	flag.Parse()

	fmt.Println(fmt.Sprintf("args: %v, %v, %v", *hostname, *consulHost, *shouldLoop))

	if *hostname == "" && *consulHost == "" {
		panic("Must provide hostname and consulhost flags")
	}

	stateUrl := ConsulStateURL{Host: *consulHost, Hostname: *hostname, Wait: *wait}
	servicesUrl := ConsulServicesURL{Host: *consulHost}

	if *shouldLoop {
		loop(&stateUrl, &servicesUrl)
	} else {
		tick(&stateUrl, &servicesUrl)
	}
}
