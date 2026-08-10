package function_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/function"
)

func FuzzComposePipeIdentities(f *testing.F) {
	for _, seed := range []int{-100, -1, 0, 1, 100} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value int) {
		inc := func(v int) int { return v + 1 }
		double := func(v int) int { return v * 2 }
		minusThree := func(v int) int { return v - 3 }
		if got, want := function.Compose(function.Identity[int], inc)(value), inc(value); got != want {
			t.Fatalf("Compose left identity = %d, want %d", got, want)
		}
		if got, want := function.Pipe(inc, function.Identity[int])(value), inc(value); got != want {
			t.Fatalf("Pipe right identity = %d, want %d", got, want)
		}
		left := function.Compose(minusThree, function.Compose(double, inc))(value)
		right := function.Compose(function.Compose(minusThree, double), inc)(value)
		if left != right {
			t.Fatalf("Compose associativity = %d, want %d", left, right)
		}
	})
}
