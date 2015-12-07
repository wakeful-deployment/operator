package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ConsulService struct {
	ID      string
	Address string
	Name    string
	Port    int
	Service string
}

func (c ConsulService) IsPresent(services []ConsulService) bool {
	for _, other := range services {
		if other.Name == c.Name {
			return true
		}
	}

	return false
}

func (c ConsulService) Register() {
	fmt.Println(fmt.Sprintf("register %v", c))
}

func (c ConsulService) Deregister() {
	fmt.Println(fmt.Sprintf("deregister %v", c))
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

	switch resp.StatusCode {
	case 200:
		return parseResponse(resp.Body)
	default:
		return nil, errors.New("Could not fetch services")
	}

	return nil, nil
}

func serviceIsWhitelisted(service ConsulService) bool {
	serviceWhiteList := []string{"consul", "statsite", "operator"}
	for _, name := range serviceWhiteList {
		if service.Name == name {
			return true
		}
	}

	return false
}

func diffServices(leftServices []ConsulService, rightServices []ConsulService) []ConsulService {
	var diff []ConsulService

	fmt.Println(fmt.Sprintf("compared %v to %v", leftServices, rightServices))

	for _, left := range leftServices {
		if serviceIsWhitelisted(left) {
			continue
		}

		// Let's assume at first it is missing
		isMissing := true

		for _, right := range rightServices {
			fmt.Println(fmt.Sprintf("comparing %s to %s", left.Name, right.Name))
			if left.Name == right.Name {
				// If we find a match, then it's not missing, but is present
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			fmt.Println(fmt.Sprintf("adding %s to the diff", left.Name))
			diff = append(diff, left)
		}
	}

	return diff
}

func normalizeConsulServices(newState ConsulState, current []ConsulService) {
	fmt.Println(fmt.Sprintf("newState: %v", newState))
	desired := newState.Services()

	removed := diffServices(current, desired)
	added := diffServices(desired, current)

	fmt.Printf("Removed services: %v\n", removed)
	fmt.Printf("Added services: %v\n", added)

	if len(added) == 0 && len(removed) == 0 {
		return
	}

	if len(added) > 0 {
		for _, service := range added {
			service.Register()
		}
	}

	if len(removed) > 0 {
		for _, service := range removed {
			service.Deregister()
		}
	}
}
