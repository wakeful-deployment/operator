package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/service"
	"net/http"
)

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

type ServiceRepresentation struct {
	Name    string
	Address string
}

func Register(s service.Service) error {
	rep := ServiceRepresentation{Name: s.Name, Address: global.Info.Consulhost}
	json, err := json.Marshal(rep)

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

func Deregister(s service.Service) error {
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

func NormalizeServices(desired []service.Service, current []service.Service) error {
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
