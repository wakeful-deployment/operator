package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/service"
	"strings"
)

func RegisteredServices(client Client) ([]service.Service, error) {
	output, err := client.RegisteredServices()

	if err != nil {
		return nil, err
	}
	services, err := parseResponse(output)

	if err != nil {
		return nil, err
	}

	return services, nil
}

func NormalizeServices(client Client, desired []service.Service, current []service.Service) error {
	removed := Diff(current, desired)
	added := Diff(desired, current)

	fmt.Printf("INFO: Removed services: %v\n", removed)
	fmt.Printf("INFO: Added services: %v\n", added)

	errs := []error{}
	for _, service := range added {
		err := client.Register(service)

		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, service := range removed {
		err := client.Deregister(service)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		errMsg := fmt.Sprintf("ERROR: At least 1 error normalizing services: %v", errs)
		return errors.New(errMsg)
	}

	return nil
}

func parseResponse(body string) ([]service.Service, error) {
	var serviceDescriptions map[string]service.Service
	var services []service.Service

	body = strings.Trim(body, "")

	if body == "" {
		return services, nil
	}

	reader := strings.NewReader(body)
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

func Diff(left []service.Service, right []service.Service) []service.Service {
	var result []service.Service

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
