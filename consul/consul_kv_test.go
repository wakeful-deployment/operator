package consul

import (
	"testing"
)

const ValidKey = "/_wakeful/nodes/bar"
const Base64Value = "Zm9vYmFy" // => foobar

var consulKV = ConsulKV{Key: ValidKey, Value: Base64Value}

func TestConsulKVToService(t *testing.T) {
	service := consulKV.ToService()

	expectedName := "bar"
	if service.Name != expectedName {
		t.Errorf("Service name = expect: %s but got: %s", expectedName, service.Name)
	}
}
