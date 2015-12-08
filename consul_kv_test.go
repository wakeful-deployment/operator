package main

import (
	"testing"
)

const ValidKey = "/_wakeful/nodes/bar"
const Base64Value = "Zm9vYmFy" // => foobar

var consulKV = ConsulKV{Key: ValidKey, Value: Base64Value}

func TestConsulKVToContainer(t *testing.T) {
	container := consulKV.ToContainer()

	expectedName := "bar"
	if container.Name != expectedName {
		t.Errorf("Container name = expect: %s but got: %s", expectedName, container.Name)
	}

	expectedImage := "foobar"
	if container.Image != expectedImage {
		t.Errorf("Container image = expect: %s but got: %s", expectedImage, container.Image)
	}
}

func TestConsulKVToService(t *testing.T) {
	service := consulKV.ToService()

	expectedName := "bar"
	if service.Name != expectedName {
		t.Errorf("Service name = expect: %s but got: %s", expectedName, service.Name)
	}
}
