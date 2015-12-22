package main

import (
	"flag"
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
	"time"
)

func tick(stateUrl *ConsulStateURL, servicesUrl *consul.ConsulServicesURL, bootstrapContainers []docker.Container) (int, error) {
	desiredState, err := NewConsulState(stateUrl.ToString())

	if err != nil {
		return 0, err
	}

	currentState, err := NewCurrentState(servicesUrl.ToString())

	if err != nil {
		return 0, err
	}

	desiredContainers := desiredState.Containers()
	desiredContainers = append(desiredContainers, bootstrapContainers...)
	currentContainers := currentState.Containers()

	err = docker.NormalizeDockerContainers(desiredContainers, currentContainers)

	if err != nil {
		fmt.Println(err)
	}

	desiredServices := desiredState.Services()
	currentServices := currentState.Services()

	err = consul.NormalizeConsulServices(desiredServices, currentServices)

	if err != nil {
		fmt.Println(err)
	}

	return desiredState.Index, nil
}

func loop(stateUrl *ConsulStateURL, servicesUrl *consul.ConsulServicesURL, bootstrapContainers []docker.Container) {
	for {
		newIndex, err := tick(stateUrl, servicesUrl, bootstrapContainers)
		stateUrl.Index = newIndex

		time.Sleep(time.Second)

		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func tickOnce(stateUrl *ConsulStateURL, servicesUrl *consul.ConsulServicesURL, bootstrapContainers []docker.Container) {
	_, err := tick(stateUrl, servicesUrl, bootstrapContainers)

	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	var (
		nodename   = flag.String("node", "", "The name of the host which is running operator")
		consulHost = flag.String("consul", "", "The name of the consul host")
		configPath = flag.String("config", "./operator.json", "The path to the operator.json. The default is the current directory of the binary.")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "5m", "The timeout for polling")
	)
	flag.Parse()

	if *nodename == "" || *consulHost == "" {
		panic("ERROR: Must provide -node and -consul flags")
	}

	consul.Node.Name = *nodename
	consul.Node.Host = *consulHost

	bootstrapContainers, err := Bootstrap(*configPath)

	if err != nil {
		panic(err)
	}

	stateUrl := ConsulStateURL{Wait: *wait}
	servicesUrl := consul.ConsulServicesURL{}

	if *shouldLoop {
		loop(&stateUrl, &servicesUrl, bootstrapContainers)
	} else {
		tickOnce(&stateUrl, &servicesUrl, bootstrapContainers)
	}
}
