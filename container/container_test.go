package container

import (
	"testing"
)

func TestDiff(t *testing.T) {
	left := []Container{
		Container{Name: "statsite"},
		Container{Name: "operator"},
		Container{Name: "consul"},
		Container{Name: "mysql"},
	}

	right := []Container{
		Container{Name: "operator"},
		Container{Name: "consul"},
		Container{Name: "redis"},
		Container{Name: "statsite"},
		Container{Name: "proxy"},
	}

	removed := Diff(left, right)
	added := Diff(right, left)

	expectedRemoved := []string{"mysql"}
	expectedAdded := []string{"redis", "proxy"}

	expectedRemovedSuccess := true
	expectedAddedSuccess := true

	for _, container := range removed {
		found := false

		for _, name := range expectedRemoved {
			if container.Name == name {
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

		for _, container := range removed {
			if container.Name == name {
				found = true
				break
			}
		}

		if !found {
			expectedRemovedSuccess = false
		}
	}

	for _, container := range added {
		found := false

		for _, name := range expectedAdded {
			if container.Name == name {
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

		for _, container := range added {
			if container.Name == name {
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
