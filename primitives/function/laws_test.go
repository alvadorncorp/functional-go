package function_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/function"
)

func TestComposeAndPipeLaws(t *testing.T) {
	t.Parallel()
	inc := func(value int) int { return value + 1 }
	double := func(value int) int { return value * 2 }
	formatLen := func(value int) string { return string(rune('a' + value%26)) }

	for _, value := range []int{-10, 0, 7, 42} {
		if got, want := function.Compose(function.Identity[int], inc)(value), inc(value); got != want {
			t.Fatalf("left Compose identity at %d = %d, want %d", value, got, want)
		}
		if got, want := function.Compose(inc, function.Identity[int])(value), inc(value); got != want {
			t.Fatalf("right Compose identity at %d = %d, want %d", value, got, want)
		}
		composeLeft := function.Compose(formatLen, function.Compose(double, inc))(value)
		composeRight := function.Compose(function.Compose(formatLen, double), inc)(value)
		if composeLeft != composeRight {
			t.Fatalf("Compose associativity at %d: %q != %q", value, composeLeft, composeRight)
		}
		if got, want := function.Pipe(function.Identity[int], inc)(value), inc(value); got != want {
			t.Fatalf("left Pipe identity at %d = %d, want %d", value, got, want)
		}
		if got, want := function.Pipe(inc, function.Identity[int])(value), inc(value); got != want {
			t.Fatalf("right Pipe identity at %d = %d, want %d", value, got, want)
		}
		pipeLeft := function.Pipe(function.Pipe(inc, double), formatLen)(value)
		pipeRight := function.Pipe(inc, function.Pipe(double, formatLen))(value)
		if pipeLeft != pipeRight {
			t.Fatalf("Pipe associativity at %d: %q != %q", value, pipeLeft, pipeRight)
		}
	}
}

func TestFlipInvolutionAndTapIdentity(t *testing.T) {
	t.Parallel()
	combine := func(a int, b string) string { return string(rune('0'+a)) + b }
	doubleFlip := function.Flip(function.Flip(combine))
	if got, want := doubleFlip(3, "x"), combine(3, "x"); got != want {
		t.Fatalf("Flip(Flip(f)) = %q, want %q", got, want)
	}
	calls := 0
	if got := function.Tap(func(int) { calls++ })(99); got != 99 || calls != 1 {
		t.Fatalf("Tap law = (%d,%d), want (99,1)", got, calls)
	}
}

func TestTransformPipeComposeLaws(t *testing.T) {
	t.Parallel()
	f := function.Transform[int](func(value int) int { return value + 1 })
	g := function.Transform[int](func(value int) int { return value * 2 })
	h := function.Transform[int](func(value int) int { return value - 3 })

	for _, value := range []int{-10, 0, 7, 42} {
		id := function.PipeTransforms[int]()
		if got, want := function.PipeTransforms(id, f)(value), f(value); got != want {
			t.Fatalf("pipe left identity at %d = %d, want %d", value, got, want)
		}
		if got, want := function.PipeTransforms(f, id)(value), f(value); got != want {
			t.Fatalf("pipe right identity at %d = %d, want %d", value, got, want)
		}
		if got, want := function.ComposeTransforms(id, f)(value), f(value); got != want {
			t.Fatalf("compose left identity at %d = %d, want %d", value, got, want)
		}
		if got, want := function.ComposeTransforms(f, id)(value), f(value); got != want {
			t.Fatalf("compose right identity at %d = %d, want %d", value, got, want)
		}

		// Associativity of PipeTransforms: (f |> g) |> h == f |> (g |> h)
		pipeLeft := function.PipeTransforms(function.PipeTransforms(f, g), h)(value)
		pipeRight := function.PipeTransforms(f, function.PipeTransforms(g, h))(value)
		if pipeLeft != pipeRight {
			t.Fatalf("PipeTransforms associativity at %d: %d != %d", value, pipeLeft, pipeRight)
		}

		// Associativity of ComposeTransforms: (f ∘ g) ∘ h == f ∘ (g ∘ h)
		composeLeft := function.ComposeTransforms(function.ComposeTransforms(f, g), h)(value)
		composeRight := function.ComposeTransforms(f, function.ComposeTransforms(g, h))(value)
		if composeLeft != composeRight {
			t.Fatalf("ComposeTransforms associativity at %d: %d != %d", value, composeLeft, composeRight)
		}

		// Pipe vs nested binary Pipe when types are homogeneous
		if got, want := function.PipeTransforms(f, g)(value), function.Pipe(f, g)(value); got != want {
			t.Fatalf("PipeTransforms vs Pipe at %d = %d, want %d", value, got, want)
		}
		if got, want := function.ComposeTransforms(f, g)(value), function.Compose(f, g)(value); got != want {
			t.Fatalf("ComposeTransforms vs Compose at %d = %d, want %d", value, got, want)
		}
	}
}
