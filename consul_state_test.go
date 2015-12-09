package main

import (
	"github.com/wakeful-deployment/operator/consul"
	"io"
	"net/http"
	"strings"
	"testing"
	// "github.com/wakeful-deployment/operator/docker"
)

type CloseReader struct {
	io.Reader
}

func (c CloseReader) Close() (err error) {
	return
}

func NewResponse(statusCode int, bodyString string, headers map[string][]string) *http.Response {
	body := CloseReader{strings.NewReader(bodyString)}
	return &http.Response{StatusCode: statusCode, Body: body, Header: headers}
}
func TestHandleConsulResponse(t *testing.T) {
	bodyString := `
		[{"CreateIndex":1714,"ModifyIndex":1714,"LockIndex":0,"Key":"_wakeful/nodes/abc/proxy","Flags":0,"Value":"Zm9vYmFy"}]
	`
	headers := make(map[string][]string)
	headers["X-Consul-Index"] = []string{"78"}
	resp := NewResponse(200, bodyString, headers)
	state := &ConsulState{}

	err := handleConsulResponse(resp, state)

	if err != nil {
		t.Fatalf("Error was non-nil: %v", err)
	}

	if index := state.Index; index != 78 {
		t.Errorf("Wrong index - expect: 78, got: %v", index)
	}

	kvs := state.KVs
	if len(kvs) != 1 {
		t.Fatalf("Wrong number of kvs - expect: 1, got: %v", len(kvs))
	}

	kv := kvs[0]

	expectedKey := "_wakeful/nodes/abc/proxy"
	if kv.Key != expectedKey {
		t.Errorf("Wrong key - expect: '%s', got: %s", expectedKey, kv.Key)
	}

	expectedValue := "Zm9vYmFy"
	if kv.Value != expectedValue {
		t.Errorf("Wrong value - expect: '%s', got: %s", expectedValue, kv.Value)
	}

	resp = NewResponse(404, "", headers)
	state = &ConsulState{}

	err = handleConsulResponse(resp, state)

	if err != nil {
		t.Fatalf("Error was non-nil: %v", err)
	}

	if index := state.Index; index != 78 {
		t.Errorf("Wrong index - expect: 78, got: %v", index)
	}

	kvs = state.KVs
	if len(kvs) != 0 {
		t.Fatalf("Wrong number of kvs - expect: 0, got: %v", len(kvs))
	}

	resp = NewResponse(500, "", make(map[string][]string))
	state = &ConsulState{}

	err = handleConsulResponse(resp, state)

	if err == nil {
		t.Fatalf("Error was nil")
	}
}

func TestStateContainers(t *testing.T) {
	kvs := []consul.ConsulKV{consul.ConsulKV{Key: "_wakeful/nodes/abc/proxy", Value: "Zm9vYmFy"}}
	state := ConsulState{Index: 29, KVs: kvs}

	containers := state.Containers()
	if len(containers) != 1 {
		t.Fatalf("Containers length was not 1")
	}

	container := containers[0]

	if container.Name != "proxy" {
		t.Errorf("Wrong name - expected: proxy, got: %s", container.Name)
	}

	if container.Image != "foobar" {
		t.Errorf("Wrong image - expected: foobar, got: %s", container.Image)
	}
}

func TestStateServices(t *testing.T) {
	kvs := []consul.ConsulKV{consul.ConsulKV{Key: "_wakeful/nodes/abc/proxy", Value: "Zm9vYmFy"}}
	state := ConsulState{Index: 29, KVs: kvs}

	services := state.Services()
	if len(services) != 1 {
		t.Fatalf("services length was not 1")
	}

	service := services[0]

	if service.Name != "proxy" {
		t.Errorf("Wrong name - expected: proxy, got: %s", service.Name)
	}
}
