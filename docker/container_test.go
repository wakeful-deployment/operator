package docker

import (
	"github.com/wakeful-deployment/operator/consul"
	"os"
	"strings"
	"testing"
)

const ValidKey = "/_wakeful/nodes/bar"
const Base64Value = "Zm9vYmFy" // => foobar

var consulKV = consul.ConsulKV{Key: ValidKey, Value: Base64Value}

func TestConsulKVToContainer(t *testing.T) {
	container := KVToContainer(consulKV)

	expectedName := "bar"
	if container.Name != expectedName {
		t.Errorf("Container name = expect: %s but got: %s", expectedName, container.Name)
	}

	expectedImage := "foobar"
	if container.Image != expectedImage {
		t.Errorf("Container image = expect: %s but got: %s", expectedImage, container.Image)
	}
}

func TestContainerPortString(t *testing.T) {
	ports := []PortPair{PortPair{Incoming: 8000, Outgoing: 8000}}
	containerWithPorts := Container{Name: "foo", Image: "foo:latest", Ports: ports}
	containerWithoutPorts := Container{Name: "foo", Image: "foo:latest"}

	expectedWithPortsString := " -p 8000:8000"
	expectedWithoutPortsString := "-P"

	if containerWithPorts.portsString() != expectedWithPortsString {
		t.Errorf("portsString is wrong, got '%s' expected '%s'", containerWithPorts.portsString(), expectedWithPortsString)
	}

	if containerWithoutPorts.portsString() != expectedWithoutPortsString {
		t.Errorf("portsString is wrong, got '%s' expected '%s'", containerWithoutPorts.portsString(), expectedWithoutPortsString)
	}
}

func TestContainerEnvString(t *testing.T) {
	os.Setenv("QUX", "myenv")
	containerWithEnv := Container{Name: "foo", Image: "foo:latest", Env: map[string]string{"foo": "bar", "baz": "$QUX", "alpha": "$omega"}}

	actualEnvString := containerWithEnv.envString()
	parts := strings.Split(actualEnvString, " ")

	correctCount := 0
	for _, part := range parts {
		if part == "" || part == "-e" {
			continue
		}

		if part == "foo=bar" || part == "baz=myenv" || part == "alpha=$omega" {
			correctCount += 1
		}
	}

	if correctCount != 3 {
		t.Errorf("envString is wrong, '%s' does not include the right vars", actualEnvString)
	}
}

func TestDockerPsParsing(t *testing.T) {
	output := "consul plum/wake-consul-server:latest\n"
	containers, err := parseDockerPsOutput(output)

	if err != nil {
		t.Fatalf("Error when parsing docker output")
	}

	if len(containers) != 1 {
		t.Fatalf("Containers length was incorrect - expect: 1, got %d", len(containers))
	}

	container := containers[0]

	if container.Name != "consul" {
		t.Errorf("Wrong container name - expected: 'consul', got: '%s'", container.Name)
	}

	if container.Image != "plum/wake-consul-server:latest" {
		t.Errorf("Wrong container name - expected: 'plum/wake-consul-server:latest', got: '%s'", container.Image)
	}
}
