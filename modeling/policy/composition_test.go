package policy_test

import (
	"strings"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/policy"
)

func TestComposeRunsFirstThenSecond(t *testing.T) {
	var trace []string
	first := policy.Policy[int, int](func(input int) int {
		trace = append(trace, "first")
		return input + 1
	})
	second := policy.Policy[int, string](func(input int) string {
		trace = append(trace, "second")
		return strings.Repeat("x", input)
	})

	if got := policy.Compose(first, second)(2); got != "xxx" {
		t.Fatalf("Compose result = %q, want xxx", got)
	}
	if got := strings.Join(trace, ","); got != "first,second" {
		t.Fatalf("Compose order = %q, want first,second", got)
	}
}

func TestComposeAssociativity(t *testing.T) {
	a := policy.Policy[int, int](func(value int) int { return value + 1 })
	b := policy.Policy[int, int](func(value int) int { return value * 2 })
	c := policy.Policy[int, string](func(value int) string { return strings.Repeat("x", value) })

	left := policy.Compose(policy.Compose(a, b), c)
	right := policy.Compose(a, policy.Compose(b, c))
	for _, input := range []int{0, 1, 4} {
		if got, want := left(input), right(input); got != want {
			t.Fatalf("Compose associativity at %d = %q, want %q", input, got, want)
		}
	}
}

func TestZipRunsLeftThenRightThenCombine(t *testing.T) {
	var trace []string
	left := policy.Policy[int, int](func(input int) int {
		trace = append(trace, "left")
		return input + 1
	})
	right := policy.Policy[int, int](func(input int) int {
		trace = append(trace, "right")
		return input * 2
	})
	zipped := policy.Zip(left, right, func(a, b int) int {
		trace = append(trace, "combine")
		return a + b
	})

	if got := zipped(3); got != 10 {
		t.Fatalf("Zip result = %d, want 10", got)
	}
	if got := strings.Join(trace, ","); got != "left,right,combine" {
		t.Fatalf("Zip order = %q, want left,right,combine", got)
	}
}

func TestZipPreservesZeroOutputs(t *testing.T) {
	left := policy.Constant[string](0)
	right := policy.Constant[string]("")
	got := policy.Zip(left, right, func(number int, text string) bool {
		return number == 0 && text == ""
	})("ignored")
	if !got {
		t.Fatal("Zip did not preserve zero outputs")
	}
}

func TestCompositionRejectsNil(t *testing.T) {
	valid := policy.Policy[int, int](twice)
	var nilPolicy policy.Policy[int, int]
	var nilCombiner func(int, int) int

	assertPanic(t, "modeling/policy: nil policy", func() {
		policy.Compose(nilPolicy, valid)
	})
	assertPanic(t, "modeling/policy: nil policy", func() {
		policy.Compose(valid, nilPolicy)
	})
	assertPanic(t, "modeling/policy: nil policy", func() {
		policy.Zip(nilPolicy, valid, func(a, b int) int { return a + b })
	})
	assertPanic(t, "modeling/policy: nil policy", func() {
		policy.Zip(valid, nilPolicy, func(a, b int) int { return a + b })
	})
	assertPanic(t, "modeling/policy: nil combiner", func() {
		policy.Zip(valid, valid, nilCombiner)
	})
}
