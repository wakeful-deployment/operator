package consul

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/wakeful-deployment/operator/service"
	"strings"
)

type KV struct {
	Key         string
	Value       string
	ModifyIndex int
}

func (kv KV) Name() string {
	keyParts := strings.Split(kv.Key, "/")
	return keyParts[len(keyParts)-1]
}

func (kv KV) DecodedValue() []byte {
	base64Value := kv.Value
	decoded, _ := base64.StdEncoding.DecodeString(base64Value)
	return decoded
}

func (kv KV) DecodeService() (*service.Service, error) {
	service := &service.Service{}
	b := kv.DecodedValue()
	reader := bytes.NewReader(b)
	err := json.NewDecoder(reader).Decode(&service)

	if err != nil {
		return nil, err
	}

	service.Name = kv.Name()

	return service, nil
}
