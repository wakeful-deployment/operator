package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func Diff(left []Service, right []Service) []Service {
	var result []Service

	for _, leftItem := range left {
		// Let's assume at first it is missing
		isMissing := true

		for _, rightItem := range right {
			if leftItem.Name() == rightItem.Name() {
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

func Register(service Service) error {
	s := S{Name_: service.Name(), Tags_: service.Tags()}

	json, err := json.Marshal(s)

	if err != nil {
		return err
	}

	reader := bytes.NewReader(json)

	resp, err := http.Post(ServiceRegisterURL(), "application/json", reader)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service failed to register: %v", s))
	}

	return nil
}

func Deregister(s Service) error {
	reader := bytes.NewReader([]byte{})
	url := ServiceDeregisterURL(s)
	resp, err := http.Post(url, "application/json", reader)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("service failed to deregister: %v", s))
	}

	return nil
}

func NormalizeServices(desired []Service, current []Service) error {
	removed := Diff(current, desired)
	added := Diff(desired, current)

	fmt.Printf("INFO: Removed services: %v\n", removed)
	fmt.Printf("INFO: Added services: %v\n", added)

	errs := []error{}
	for _, service := range added {
		err := Register(service)

		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, service := range removed {
		err := Deregister(service)
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
