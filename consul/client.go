package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/service"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client interface {
	Register(service.Service) error
	Deregister(service.Service) error
	RegisteredServices() (string, error)
	PostMetadata(string, map[string]string) error
	Detect() error
	GetDirectoryState(string, int, string) (*DirectoryState, error)
	ConsulHost() string
}

type HttpClient struct {
	Host string
}

func (h HttpClient) ConsulHost() string {
	return h.Host
}

func (h HttpClient) Register(s service.Service) error {
	rep := ServiceRepresentation{Name: s.Name, Address: h.ConsulHost()}
	json, err := json.Marshal(rep)

	if err != nil {
		return err
	}

	reader := bytes.NewReader(json)

	resp, err := http.Post(h.serviceRegisterURL(), "application/json", reader)

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
	url := h.serviceDeregisterURL(s)
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
	resp, err := http.Get(h.servicesURL())

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
		request, err := http.NewRequest("PUT", h.metadataURL(key, nodeName), strings.NewReader(value))

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

func (h HttpClient) GetDirectoryState(nodeName string, index int, wait string) (*DirectoryState, error) {
	url := h.directoryStateURL(nodeName, index, wait)
	state := &DirectoryState{}
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	err = handleDirectoryResponse(resp, state)

	if err != nil {
		return nil, err
	}

	return state, nil
}

func getDirectoryIndex(resp *http.Response) (int, error) {
	return strconv.Atoi(resp.Header["X-Consul-Index"][0])
}

func handleDirectoryResponse(resp *http.Response, state *DirectoryState) error {
	switch resp.StatusCode {
	case 200:
		index, err := getDirectoryIndex(resp)

		if err != nil {
			return err
		}

		var keys []KV
		err = json.NewDecoder(resp.Body).Decode(&keys)

		if err != nil {
			return err
		}

		state.KVs = keys
		state.Index = index

		return nil
	case 404:
		index, err := getDirectoryIndex(resp)

		if err != nil {
			return err
		}

		state.Index = index

		return nil
	default:
		return errors.New("non 200/404 response code")
	}
}

const consulCheckTimeout = 5 * time.Second

func (h HttpClient) Detect() error {
	url := h.consulCheckURL()

	client := http.Client{Timeout: consulCheckTimeout}
	resp, err := client.Get(url)

	if err == nil && resp.StatusCode == 200 {
		return nil
	}

	if err != nil {
		return err
	}

	return errors.New(fmt.Sprintf("consul check failed with non-200 response: %d", resp.StatusCode))
}

type ServiceRepresentation struct {
	Name    string
	Address string
}

func (h HttpClient) consulCheckURL() string {
	return fmt.Sprintf("http://%s:8500/", h.ConsulHost)
}

func (h HttpClient) servicesURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/services", h.ConsulHost)
}

func (h HttpClient) serviceRegisterURL() string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/register", h.ConsulHost)
}

func (h HttpClient) serviceDeregisterURL(s service.Service) string {
	return fmt.Sprintf("http://%s:8500/v1/agent/service/deregister/%s", h.ConsulHost, s.Name)
}

func (h HttpClient) metadataURL(key string, nodeName string) string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s/metadata/%s", h.ConsulHost, nodeName, key)
}

func (h HttpClient) directoryStateURL(nodeName string, index int, wait string) string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s/apps/?recurse=true&index=%d&wait=%s", h.ConsulHost, nodeName, index, wait)
}
