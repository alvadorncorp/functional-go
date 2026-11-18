package function_test

import (
	"testing"

	"go.alvadorn.org/functional/primitives/function"
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

func FuzzPipeComposeTransforms(f *testing.F) {
	for _, seed := range []int{-100, -1, 0, 1, 100} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value int) {
		inc := function.Transform[int](func(v int) int { return v + 1 })
		double := function.Transform[int](func(v int) int { return v * 2 })
		minusThree := function.Transform[int](func(v int) int { return v - 3 })
		if got, want := function.PipeTransforms(function.PipeTransforms[int](), inc)(value), inc(value); got != want {
			t.Fatalf("PipeTransforms empty identity = %d, want %d", got, want)
		}
		left := function.PipeTransforms(function.PipeTransforms(inc, double), minusThree)(value)
		right := function.PipeTransforms(inc, function.PipeTransforms(double, minusThree))(value)
		if left != right {
			t.Fatalf("PipeTransforms associativity = %d, want %d", left, right)
		}
		composeLeft := function.ComposeTransforms(function.ComposeTransforms(minusThree, double), inc)(value)
		composeRight := function.ComposeTransforms(minusThree, function.ComposeTransforms(double, inc))(value)
		if composeLeft != composeRight {
			t.Fatalf("ComposeTransforms associativity = %d, want %d", composeLeft, composeRight)
		}
	})
}
