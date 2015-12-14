package docker

import (
	"os"
	"testing"
)

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

	expectedEnvString := " -e foo=bar -e baz=myenv -e alpha=$omega"
	actualEnvString := containerWithEnv.envString()

	if actualEnvString != expectedEnvString {
		t.Errorf("envString is wrong, got '%s' expected '%s'", actualEnvString, expectedEnvString)
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
