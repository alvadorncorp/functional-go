package policy_test

import (
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/policy"
)

func FuzzPolicyCompositionLaws(f *testing.F) {
	for _, seed := range []int{-10, -1, 0, 1, 10} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input int) {
		a := policy.Policy[int, int](func(value int) int { return value + 1 })
		b := policy.Policy[int, int](func(value int) int { return value * 2 })
		c := policy.Policy[int, string](strconv.Itoa)

		left := policy.Compose(policy.Compose(a, b), c)
		right := policy.Compose(a, policy.Compose(b, c))
		if got, want := left(input), right(input); got != want {
			t.Fatalf("Compose associativity at %d = %q, want %q", input, got, want)
		}
	})
}
