package test

import (
	"github.com/wakeful-deployment/operator/consul"
	"github.com/wakeful-deployment/operator/service"
)

type ConsulClient struct {
	RegisteredServicesFunction func() (string, error)
	RegisterResponse           error
	DeregisterResponse         error
	PostMetadataResponse       error
	DetectResponse             error
	GetDirectoryStateFunction  func() (*consul.DirectoryState, error)
	ConsulHostValue            string
}

func (t ConsulClient) RegisteredServices() (string, error) {
	result, err := t.RegisteredServicesFunction()

	if err != nil {
		return "", err
	}

	return result, nil
}

func (t ConsulClient) Register(s service.Service) error {
	return t.RegisterResponse
}

func (t ConsulClient) Deregister(s service.Service) error {
	return t.DeregisterResponse
}

func (t ConsulClient) PostMetadata(nodeName string, data map[string]string) error {
	return t.PostMetadataResponse
}

func (t ConsulClient) Detect() error {
	return t.DetectResponse
}

func (t ConsulClient) GetDirectoryState(nodeName string, index int, wait string) (*consul.DirectoryState, error) {
	result, err := t.GetDirectoryStateFunction()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t ConsulClient) ConsulHost() string {
	return t.ConsulHostValue
}
