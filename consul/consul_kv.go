package consul

import (
	"encoding/base64"
	"strings"
)

type ConsulKV struct {
	Key         string
	Value       string
	ModifyIndex int
}

func (c ConsulKV) Name() string {
	keyParts := strings.Split(c.Key, "/")
	return keyParts[len(keyParts)-1]
}

func (c ConsulKV) DecodedValue() string {
	base64Value := c.Value
	decoded, _ := base64.StdEncoding.DecodeString(base64Value)
	return string(decoded)
}

func (c ConsulKV) ImageName() string {
	return c.DecodedValue()
}

func (c ConsulKV) ToService() ConsulService {
	return ConsulService{ID: c.Name(), Name: c.Name()}
}
