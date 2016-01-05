package service

import (
	"testing"
)

func TestSimplePorts(t *testing.T) {
	s := Service{
		Ports: []PortPair{
			PortPair{Incoming: 8000, Outgoing: 9000},
			PortPair{Incoming: 8300, Outgoing: 9300, UDP: true},
		},
	}

	ports := s.SimplePorts()

	if len(ports) != 2 {
		t.Fatal("did not return the correct amount of ports - cannot continue")
	}

	firstPort := ports[0]
	secondPort := ports[1]

	expectedFirst := "8000:9000"
	if firstPort != expectedFirst {
		t.Errorf("expected %s, but got %s", expectedFirst, firstPort)
	}

	expectedSecond := "8300:9300/udp"
	if secondPort != expectedSecond {
		t.Errorf("expected %s, but got %s", expectedSecond, secondPort)
	}
}

func TestFullEnv(t *testing.T) {
	s := Service{
		Env: map[string]string{
			"FOO": "BAR",
		},
	}

	env := s.FullEnv()

	expectedKeys := []string{
		"FOO",
		"SERVICENAME",
		"NODE",
		"CONSULHOST",
	}

	success := true

	for _, key := range expectedKeys {
		_, prs := env[key]

		if !prs {
			success = false
			break
		}
	}

	for key, _ := range env {
		found := false

		for _, k := range expectedKeys {
			if k == key {
				found = true
			}
		}

		if !found {
			success = false
		}
	}

	if !success {
		t.Errorf("expected %v, but got %v", expectedKeys, env)
	}
}

func TestContainer(t *testing.T) {
	s := Service{
		Name:  "redis",
		Image: "redis",
	}

	expectedName := "redis"
	expectedImage := "redis"

	c := s.Container()

	if c.Name != expectedName {
		t.Errorf("expected Name to be %s, but was %s", expectedName, c.Name)
	}

	if c.Image != expectedImage {
		t.Errorf("expected Image to be %s, but was %s", expectedImage, c.Image)
	}
}

func TestDiff(t *testing.T) {
	left := []Service{
		Service{Name: "statsite"},
		Service{Name: "operator"},
		Service{Name: "consul"},
		Service{Name: "mysql"},
	}

	right := []Service{
		Service{Name: "operator"},
		Service{Name: "consul"},
		Service{Name: "redis"},
		Service{Name: "statsite"},
		Service{Name: "proxy"},
	}

	removed := Diff(left, right)
	added := Diff(right, left)

	expectedRemoved := []string{"mysql"}
	expectedAdded := []string{"redis", "proxy"}

	expectedRemovedSuccess := true
	expectedAddedSuccess := true

	for _, service := range removed {
		found := false

		for _, name := range expectedRemoved {
			if service.Name == name {
				found = true
				break
			}
		}

		if !found {
			expectedRemovedSuccess = false
		}
	}

	for _, name := range expectedRemoved {
		found := false

		for _, service := range removed {
			if service.Name == name {
				found = true
				break
			}
		}

		if !found {
			expectedRemovedSuccess = false
		}
	}

	for _, service := range added {
		found := false

		for _, name := range expectedAdded {
			if service.Name == name {
				found = true
				break
			}
		}

		if !found {
			expectedAddedSuccess = false
		}
	}

	for _, name := range expectedAdded {
		found := false

		for _, service := range added {
			if service.Name == name {
				found = true
				break
			}
		}

		if !found {
			expectedAddedSuccess = false
		}
	}

	if !expectedRemovedSuccess {
		t.Errorf("expected removed to be %v, but was %v", expectedRemoved, removed)
	}

	if !expectedAddedSuccess {
		t.Errorf("expected added to be %v, but was %v", expectedAdded, added)
	}
}
