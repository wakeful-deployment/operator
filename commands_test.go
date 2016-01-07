package main

import (
	// "github.com/wakeful-deployment/operator/test"
	"github.com/wakeful-deployment/operator/global"
	"io/ioutil"
	"testing"
)

func TestSuccessfulLoadBootStateFromFile(t *testing.T) {
	global.Machine.ForceTransition(global.Booting, nil)
	defer global.Machine.ForceTransition(global.Initial, nil)

	f, err := ioutil.TempFile("", "operator-json")

	if err != nil {
		t.Fatal("Couldn't create a tmp file for this test")
	}

	_, err = f.WriteString(`{
	"metadata": {
		"foo": "bar"
	},
	"services": {
		"statsite": {
		"image": "wakeful/wake-statsite:latest",
		"ports": [{
			"incoming": 8125,
			"outgoing": 8125,
			"udp": true
		}],
		"env": {},
		"restart": "always",
		"tags": ["statsd", "udp"],
		"checks": []
	  }
	}
}`)

	if err != nil {
		t.Fatal("Couldn't write to the tmp file")
	}

	state := LoadBootStateFromFile(f.Name())

	m := state.Metadata
	lenM := len(m)

	if lenM != 1 {
		t.Errorf("Expected length of metadata to be 1, but got %d", lenM)
	}

	fooValue := m["foo"]
	if fooValue != "bar" {
		t.Errorf("Expected [foo] to be bar, but got %s", fooValue)
	}

	lenServices := len(state.Services)

	if lenServices != 1 {
		t.Fatalf("Expected 1 service, but got %d", lenServices)
	}

	s := state.Services["statsite"]

	if s.Name != "statsite" {
		t.Errorf("Expected service name to be statsite, but got %s", s.Name)
	}

	expectedImage := "wakeful/wake-statsite:latest"
	if s.Image != expectedImage {
		t.Errorf("Expected image to be %s, but got %s", expectedImage, s.Image)
	}
}
