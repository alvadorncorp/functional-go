package function_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/function"
)

func TestCoreCombinators(t *testing.T) {
	t.Parallel()

	if got := function.Identity("value"); got != "value" {
		t.Fatalf("Identity = %q, want value", got)
	}
	if got := function.Constant[int, string](7)("ignored"); got != 7 {
		t.Fatalf("Constant = %d, want 7", got)
	}
	length := func(value string) int { return len(value) }
	double := func(value int) int { return value * 2 }
	if got := function.Compose(double, length)("abcd"); got != 8 {
		t.Fatalf("Compose = %d, want 8", got)
	}
	if got := function.Pipe(length, double)("abcd"); got != 8 {
		t.Fatalf("Pipe = %d, want 8", got)
	}
	subtract := func(a, b int) int { return a - b }
	if got := function.Flip(subtract)(2, 7); got != 5 {
		t.Fatalf("Flip = %d, want 5", got)
	}
	calls, observed := 0, 0
	tapped := function.Tap(func(value int) { calls++; observed = value })
	if got := tapped(11); got != 11 || calls != 1 || observed != 11 {
		t.Fatalf("Tap = (%d, calls=%d, observed=%d), want (11,1,11)", got, calls, observed)
	}
}
