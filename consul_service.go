package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ConsulServicesURL struct {
	Host string
}

func (c *ConsulServicesURL) ToString() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/services", c.Host)
}

type ConsulService struct {
	ID      string
	Address string
	Name    string
	Port    int
	Check   ConsulCheck
}

type ConsulCheck struct {
	HTTP     string
	Interval string
	TTL      string
}

func (c ConsulService) GetName() string {
	return c.Name
}

func (c ConsulService) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c ConsulService) Register(host string) error {
	json, err := c.ToJSON()
	reader := bytes.NewReader(json)

	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:8500/v1/agent/service/register", host)
	resp, respErr := http.Post(url, "application/json", reader)

	if respErr != nil {
		return respErr
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service failed to register: %v", c))
	}

	return nil
}

func (c ConsulService) Deregister(host string) error {
	reader := bytes.NewReader([]byte{})
	url := fmt.Sprintf("http://%s:8500/v1/agent/service/deregister/%s", host, c.ID)
	resp, respErr := http.Post(url, "application/json", reader)

	if respErr != nil {
		return respErr
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service failed to deregister: %v", c))
	}

	return nil
}

func DefaultCheck() ConsulCheck {
	return ConsulCheck{HTTP: "http://localhost:8000/_health", Interval: "6s", TTL: "5s"}
}

func parseResponse(reader io.Reader) ([]ConsulService, error) {
	var serviceDescriptions map[string]ConsulService
	var services []ConsulService

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

func GetConsulServices(url ConsulServicesURL) ([]ConsulService, error) {
	resp, err := http.Get(url.ToString())

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		return parseResponse(resp.Body)
	} else {
		return nil, errors.New("Could not fetch services")
	}
}

func diffServices(leftServices []ConsulService, rightServices []ConsulService) []ConsulService {
	var diff []ConsulService

	for _, left := range leftServices {
		if isWhitelisted(left) {
			continue
		}

		// Let's assume at first it is missing
		isMissing := true

		for _, right := range rightServices {
			if left.Name == right.Name {
				// If we find a match, then it's not missing
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			diff = append(diff, left)
		}
	}

	return diff
}

func normalizeConsulServices(newState ConsulState, current []ConsulService, consulHost string) {
	desired := newState.Services()

	removed := diffServices(current, desired)
	added := diffServices(desired, current)

	fmt.Printf("Removed services: %v\n", removed)
	fmt.Printf("Added services: %v\n", added)

	for _, service := range added {
		service.Check = DefaultCheck()
		service.Register(consulHost)
	}

	for _, service := range removed {
		service.Deregister(consulHost)
	}
}