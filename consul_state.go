package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type ConsulState struct {
	Index int
	KVs   []ConsulKV
}

func (c ConsulState) Containers() []Container {
	var containers []Container

	for _, kv := range c.KVs {
		containers = append(containers, kv.ToContainer())
	}

	return containers
}

func (c ConsulState) Services() []ConsulService {
	var services []ConsulService

	for _, kv := range c.KVs {
		services = append(services, kv.ToService())
	}

	return services
}

func getIndex(resp *http.Response) (int, error) {
	return strconv.Atoi(resp.Header["X-Consul-Index"][0])
}

func GetConsulState(state *ConsulState, url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		index, err := getIndex(resp)

		if err != nil {
			return err
		}

		var keys []ConsulKV
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
		state.KVs = []ConsulKV{}

		return nil
	default:
		return errors.New("non 200/404 response code")
	}
}
