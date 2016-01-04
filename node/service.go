package node

import (
	"fmt"
)

type PortPair struct {
	Incoming int  `json:"incoming"`
	Outgoing int  `json:"outgoing"`
	UDP      bool `json:"udp"`
}

// type Check struct{}
// func NewHTTPHealthCheck(...) HealthCheck { ... return with sane defaults }

type ServicesMap map[string]Service

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
		str := fmt.Sprintf("%s:%s", pair.Incoming, pair.Outgoing)

		if pair.UDP {
			str = fmt.Sprintf("%s/udp", str)
		}

		ports = append(ports, str)
	}

	return ports
}

func (s Service) FullEnv() map[string]string {
	var env map[string]string

	for key, value := range s.Env {
		env[key] = value
	}

	env["SERVICENAME"] = s.Name
	env["NODE"] = Info.Name
	env["CONSULHOST"] = Info.Host

	return env
}

func (s Service) Container() Container {
	return Container{
		Name_:    s.Name,
		Image_:   s.Image,
		Ports_:   s.SimplePorts(),
		Env_:     s.FullEnv(),
		Restart_: s.Restart,
		Tags_:    s.Tags,
	}
}

func (s Service) ConsulService() ConsulService {
	return ConsulService{
		Name_: s.Name,
		Tags_: s.Tags,
	}
}
