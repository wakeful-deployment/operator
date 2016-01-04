package directory

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
	Service     service.Service
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

func DecodeService(kv KV) error {
	service := service.Service{}
	b := kv.DecodedValue()
	reader := bytes.NewReader(b)
	err := json.NewDecoder(reader).Decode(&service)

	if err != nil {
		return err
	}

	service.Name = kv.Name()
	kv.Service = service

	return nil
}
