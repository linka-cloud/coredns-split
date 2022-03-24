package split

import (
	"testing"

	"github.com/coredns/caddy"
)

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `split`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
}
