package consul

import (
	"github.com/wakeful-deployment/operator/service"
)

type DirectoryState struct {
	Index int
	KVs   []KV
}

func (s DirectoryState) Services() ([]*service.Service, error) {
	var services []*service.Service

	for _, kv := range s.KVs {
		service, err := kv.DecodeService()

		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}
