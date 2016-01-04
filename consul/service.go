package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/node"
	"io"
	"net/http"
)

func ServicesURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/services", node.Info.Host)
}

func ServiceRegisterURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/register", node.Info.Host)
}

func ServiceDeregisterURL(s Service) string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/deregister/%s", node.Info.Host, s.Name())
}

type Service interface {
	Name() string
	Tags() []string
}

func parseResponse(reader io.Reader) ([]S, error) {
	var serviceDescriptions map[string]S
	var services []S

	err := json.NewDecoder(reader).Decode(&serviceDescriptions)

	if err != nil {
		return nil, err
	}

	for name, service := range serviceDescriptions {
		service.Name_ = name
		services = append(services, service)
	}

	return services, nil
}

func RegisteredServices() ([]S, error) {
	resp, err := http.Get(ServicesURL())

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		return parseResponse(resp.Body)
	} else {
		return nil, errors.New("Could not fetch services")
	}
}
