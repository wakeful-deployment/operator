package main

import (
	"testing"
)

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
