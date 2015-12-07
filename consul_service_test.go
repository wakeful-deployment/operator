package main

import (
	"strings"
	"testing"
)

const ConsulServicesReturnString = "{\"consul\":{\"ID\":\"consul\",\"Service\":\"consul\",\"Tags\":[],\"Address\":\"\",\"Port\":8300}}"

func TestConsulParseJSON(t *testing.T) {
	reader := strings.NewReader(ConsulServicesReturnString)
	services, err := parseResponse(reader)

	if err != nil {
		t.Errorf("%v", err)
	}

	if len(services) != 1 {
		t.Error("len(services) was not 1")
	} else if len(services) > 0 {
		firstService := services[0]
		expectedName := "consul"

		if firstService.ID != expectedName {
			t.Errorf("Expected service ID to be %s not %s", expectedName, firstService.ID)
		}

		if firstService.Name != expectedName {
			t.Errorf("Expected service name to be %s not %s", expectedName, firstService.Name)
		}
	}
}
