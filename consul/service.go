package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/service"
	"io"
	"net/http"
)

func ServicesURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/services", global.Info.Consulhost)
}

func ServiceRegisterURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/register", global.Info.Consulhost)
}

func ServiceDeregisterURL(s service.Service) string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/deregister/%s", global.Info.Consulhost, s.Name)
}

func parseResponse(reader io.Reader) ([]service.Service, error) {
	var serviceDescriptions map[string]service.Service
	var services []service.Service

	err := json.NewDecoder(reader).Decode(&serviceDescriptions)

	if err != nil {
		return nil, err
	}

	for name, service := range serviceDescriptions {
		service.Name = name
		services = append(services, service)
	}

	return services, nil
}

func RegisteredServices() ([]service.Service, error) {
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
