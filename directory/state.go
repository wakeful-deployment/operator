package directory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wakeful-deployment/operator/global"
	"github.com/wakeful-deployment/operator/service"
	"net/http"
	"strconv"
)

type StateURL struct {
	Index int
	Wait  string
}

func (c *StateURL) String() string {
	return fmt.Sprintf("http://%s:8500/v1/kv/_wakeful/nodes/%s/apps/?recurse=true&index=%d&wait=%s", global.Info.Consulhost, global.Info.Nodename, c.Index, c.Wait)
}

type State struct {
	Index int
	KVs   []KV
}

func (s State) Services() ([]service.Service, error) {
	var services []service.Service

	for _, kv := range s.KVs {
		err := DecodeService(kv)

		if err != nil {
			services = append(services, kv.Service)
		} else {
			return nil, err
		}
	}

	return services, nil
}

func getIndex(resp *http.Response) (int, error) {
	return strconv.Atoi(resp.Header["X-Consul-Index"][0])
}

func handleConsulResponse(resp *http.Response, state *State) error {
	switch resp.StatusCode {
	case 200:
		index, err := getIndex(resp)

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
		index, err := getIndex(resp)

		if err != nil {
			return err
		}

		state.Index = index

		return nil
	default:
		return errors.New("non 200/404 response code")
	}
}

func decodeServices(kvs []KV) error {
	for _, kv := range kvs {
		err := DecodeService(kv)

		if err != nil {
			return err
		}
	}

	return nil
}

func DesiredState(url string) (*State, error) {
	state := &State{}
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	err = handleConsulResponse(resp, state)

	if err != nil {
		return nil, err
	}

	err = decodeServices(state.KVs)

	if err != nil {
		return nil, err
	}

	return state, nil
}
