package policy_test

import (
	"fmt"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/policy"
)

var _ policy.Policy[int, string] = func(input int) string {
	return fmt.Sprintf("value:%d", input)
}

func TestPolicyAcceptsOrdinaryFunction(t *testing.T) {
	var p policy.Policy[int, int] = func(input int) int { return input * 2 }
	if got := p(21); got != 42 {
		t.Fatalf("policy result = %d, want 42", got)
	}
}

func TestConstantIgnoresInput(t *testing.T) {
	p := policy.Constant[int]("fixed")
	if got := p(1); got != "fixed" {
		t.Fatalf("first result = %q, want fixed", got)
	}
	if got := p(999); got != "fixed" {
		t.Fatalf("second result = %q, want fixed", got)
	}
}

func TestConstantPreservesZeroValue(t *testing.T) {
	p := policy.Constant[string](0)
	if got := p("ignored"); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func assertPanic(t *testing.T, want string, run func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	run()
}
