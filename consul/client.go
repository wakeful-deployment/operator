package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/service"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client interface {
	Register(service.Service) error
	Deregister(service.Service) error
	RegisteredServices() (string, error)
	PostMetadata(string, map[string]string) error
}

type HttpClient struct{}

func (h HttpClient) Register(s service.Service) error {
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

func (h HttpClient) Deregister(s service.Service) error {
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

func (h HttpClient) RegisteredServices() (string, error) {
	resp, err := http.Get(ServicesURL())

	if err != nil {
		return "", err
	}

	if resp.StatusCode == 200 {
		reader := resp.Body
		defer reader.Close()
		contents, err := ioutil.ReadAll(reader)

		if err != nil {
			return "", err
		}

		return string(contents), nil
	} else {
		return "", errors.New("Could not fetch services")
	}
}

func (h HttpClient) PostMetadata(nodeName string, metadata map[string]string) error {
	for key, value := range metadata {
		client := &http.Client{}
		request, err := http.NewRequest("PUT", MetadataURL(key, nodeName), strings.NewReader(value))

		if err != nil {
			return err
		}

		resp, err := client.Do(request)

		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Metadata request return non-200 response: %d", resp.StatusCode))
		}
	}

	return nil
}

type ServiceRepresentation struct {
	Name    string
	Address string
}

func ServicesURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/services", global.Info.Consulhost)
}

func ServiceRegisterURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/register", global.Info.Consulhost)
}

func ServiceDeregisterURL(s service.Service) string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/deregister/%s", global.Info.Consulhost, s.Name)
}

func MetadataURL(key string, nodeName string) string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s/metadata/%s", global.Info.Consulhost, nodeName, key)
}
