package version

import "testing"

func TestVersion(t *testing.T) {
	if Version != "0.1.0" {
		t.Fatalf("got %q", Version)
	}
}
