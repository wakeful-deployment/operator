package test

import (
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/service"
)

type ConsulClient struct {
	RegisteredServicesResponse func() (string, error)
	RegisterResponse           func(service.Service) error
	DeregisterResponse         func(service.Service) error
	PostMetadataResponse       func() error
	DetectResponse             func() error
	GetDirectoryStateResponse  func() (*consul.DirectoryState, error)
	ConsulHostResponse         func() string
}

func (t ConsulClient) RegisteredServices() (string, error) {
	result, err := t.RegisteredServicesResponse()

	if err != nil {
		return "", err
	}

	return result, nil
}

func (t ConsulClient) Register(s service.Service) error {
	return t.RegisterResponse(s)
}

func (t ConsulClient) Deregister(s service.Service) error {
	return t.DeregisterResponse(s)
}

func (t ConsulClient) PostMetadata(nodeName string, data map[string]string) error {
	return t.PostMetadataResponse()
}

func (t ConsulClient) Detect() error {
	return t.DetectResponse()
}

func (t ConsulClient) GetDirectoryState(nodeName string, index int, wait string) (*consul.DirectoryState, error) {
	result, err := t.GetDirectoryStateResponse()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t ConsulClient) ConsulHost() string {
	return t.ConsulHostResponse()
}
