package tool_test

import (
	"testing"
)

func TestNoneFallback(t *testing.T) {
	if got := noneFallback("   "); got != "none" {
		t.Fatalf("expected none for empty input, got %q", got)
	}
	if got := noneFallback("  x  "); got != "x" {
		t.Fatalf("expected trimmed value, got %q", got)
	}
}
