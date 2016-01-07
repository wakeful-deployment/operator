package service

import (
	"fmt"
	"github.com/wakeful-deployment/operator/container"
	"github.com/wakeful-deployment/operator/global"
)

func Diff(left []Service, right []Service) []Service {
	var result []Service

	for _, leftItem := range left {
		// Let's assume at first it is missing
		isMissing := true

		for _, rightItem := range right {
			if leftItem.Name == rightItem.Name {
				// If we find a match, then it's not missing
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			result = append(result, leftItem)
		}
	}

	return result
}

type PortPair struct {
	Incoming int  `json:"incoming"`
	Outgoing int  `json:"outgoing"`
	UDP      bool `json:"udp"`
}

// type Check struct{}
// func NewHTTPHealthCheck(...) HealthCheck { ... return with sane defaults }

type Service struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Ports   []PortPair        `json:"ports"`
	Env     map[string]string `json:"env"`
	Restart string            `json:"restart"`
	Tags    []string          `json:"tags"`
	// Checks  []Check           `json:"checks"`
}

func (s Service) SimplePorts() []string {
	var ports []string

	for _, pair := range s.Ports {
		str := fmt.Sprintf("%d:%d", pair.Incoming, pair.Outgoing)

		if pair.UDP {
			str = fmt.Sprintf("%s/udp", str)
		}

		ports = append(ports, str)
	}

	return ports
}

func (s Service) FullEnv(nodeName string) map[string]string {
	env := make(map[string]string)

	for key, value := range s.Env {
		env[key] = value
	}

	env["SERVICENAME"] = s.Name
	env["NODE"] = nodeName
	env["CONSULHOST"] = global.Info.Consulhost

	return env
}

func (s Service) Container(nodeName string) container.Container {
	return container.Container{
		Name:    s.Name,
		Image:   s.Image,
		Ports:   s.SimplePorts(),
		Env:     s.FullEnv(nodeName),
		Restart: s.Restart,
		Tags:    s.Tags,
	}
}
