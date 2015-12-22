package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/docker"
	"net/http"
	"strconv"
)

type ConsulStateURL struct {
	Index int
	Wait  string
}

func (c *ConsulStateURL) ToString() string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s?recurse=true&index=%d&wait=%s", consul.Node.Host, consul.Node.Name, c.Index, c.Wait)
}

type ConsulState struct {
	Index int
	KVs   []consul.ConsulKV
}

func (c ConsulState) Containers() []docker.Container {
	var containers []docker.Container

	for _, kv := range c.KVs {
		containers = append(containers, docker.KVToContainer(kv))
	}

	return containers
}

func (c ConsulState) Services() []consul.ConsulService {
	var services []consul.ConsulService

	for _, kv := range c.KVs {
		services = append(services, kv.ToService())
	}

	return services
}

func getIndex(resp *http.Response) (int, error) {
	return strconv.Atoi(resp.Header["X-Consul-Index"][0])
}

func handleConsulResponse(resp *http.Response, state *ConsulState) error {
	switch resp.StatusCode {
	case 200:
		index, err := getIndex(resp)

		if err != nil {
			return err
		}

		var keys []consul.ConsulKV
		err = json.NewDecoder(resp.Body).Decode(&keys)

		if err != nil {
			return err
		}

		state.KVs = keys
		state.Index = index

		return nil
	case 404:
		index, err := getIndex(resp)

		if err != nil {
			return err
		}

		state.Index = index
		state.KVs = []consul.ConsulKV{}

		return nil
	default:
		return errors.New("non 200/404 response code")
	}
}

func NewConsulState(url string) (*ConsulState, error) {
	state := &ConsulState{}
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	err = handleConsulResponse(resp, state)

	if err != nil {
		return nil, err
	}

	return state, nil
}
