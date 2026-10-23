package function_test

import (
	"strings"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/function"
)

type user struct {
	Name   string
	Active bool
}

func TestTransformDirectInvocation(t *testing.T) {
	t.Parallel()

	upper := function.Transform[string](strings.ToUpper)
	if got, want := upper("go"), "GO"; got != want {
		t.Fatalf("Transform(string) = %q, want %q", got, want)
	}

	inc := function.Transform[int](func(value int) int { return value + 1 })
	if got, want := inc(41), 42; got != want {
		t.Fatalf("Transform(int) = %d, want %d", got, want)
	}

	activate := function.Transform[user](func(value user) user {
		value.Active = true
		return value
	})
	if got := activate(user{Name: "a"}); !got.Active || got.Name != "a" {
		t.Fatalf("Transform(user) = %+v, want Active with Name a", got)
	}

	ptr := function.Transform[*int](func(value *int) *int {
		if value != nil {
			next := *value + 1
			return &next
		}
		return value
	})
	seed := 7
	if got := ptr(&seed); got == nil || *got != 8 {
		t.Fatalf("Transform(*int) = %v, want pointer to 8", got)
	}
}

func TestPipeTransformsEmptyIsIdentity(t *testing.T) {
	t.Parallel()
	transform := function.PipeTransforms[int]()
	for _, value := range []int{-3, 0, 9} {
		if got := transform(value); got != value {
			t.Fatalf("PipeTransforms()(%d) = %d, want %d", value, got, value)
		}
	}
}

func TestPipeTransformsSingleIsEquivalent(t *testing.T) {
	t.Parallel()
	double := function.Transform[int](func(value int) int { return value * 2 })
	piped := function.PipeTransforms(double)
	if got, want := piped(21), double(21); got != want {
		t.Fatalf("PipeTransforms(f) = %d, want %d", got, want)
	}
}

func TestPipeTransformsOrderLeftToRight(t *testing.T) {
	t.Parallel()
	// 5 +10 → 15 ×2 → 30
	inc := function.Transform[int](func(value int) int { return value + 10 })
	double := function.Transform[int](func(value int) int { return value * 2 })
	transform := function.PipeTransforms(inc, double)
	if got, want := transform(5), 30; got != want {
		t.Fatalf("PipeTransforms order = %d, want %d", got, want)
	}
}

func TestComposeTransformsEmptyIsIdentity(t *testing.T) {
	t.Parallel()
	transform := function.ComposeTransforms[string]()
	if got, want := transform("x"), "x"; got != want {
		t.Fatalf("ComposeTransforms() = %q, want %q", got, want)
	}
}

func TestComposeTransformsOrderRightToLeft(t *testing.T) {
	t.Parallel()
	// Compose(f,g,h)(x) = f(g(h(x)))
	// h: +1, g: *2, f: +10 → ((5+1)*2)+10 = 22
	addOne := function.Transform[int](func(value int) int { return value + 1 })
	double := function.Transform[int](func(value int) int { return value * 2 })
	addTen := function.Transform[int](func(value int) int { return value + 10 })
	transform := function.ComposeTransforms(addTen, double, addOne)
	if got, want := transform(5), 22; got != want {
		t.Fatalf("ComposeTransforms order = %d, want %d", got, want)
	}
}

func TestPipeTransformsClosureUnderComposition(t *testing.T) {
	t.Parallel()
	rename := function.Transform[user](func(value user) user {
		value.Name = "Alice"
		return value
	})
	activate := function.Transform[user](func(value user) user {
		value.Active = true
		return value
	})
	// PipeTransforms returns Transform[T], proving closure under composition.
	transform := function.PipeTransforms(rename, activate)
	got := transform(user{})
	if got.Name != "Alice" || !got.Active {
		t.Fatalf("closure result = %+v, want Alice/active", got)
	}
}

func TestPipeTransformsCopiesVariadicSlice(t *testing.T) {
	t.Parallel()
	inc := function.Transform[int](func(value int) int { return value + 10 })
	double := function.Transform[int](func(value int) int { return value * 2 })
	steps := []function.Transform[int]{inc, double}
	transform := function.PipeTransforms(steps...)
	steps[0] = nil
	steps[1] = func(value int) int { return -value }
	// 5 +10 → 15 ×2 → 30; an aliased slice would panic on nil or yield -5.
	if got, want := transform(5), 30; got != want {
		t.Fatalf("PipeTransforms after caller mutation = %d, want %d", got, want)
	}
}

func TestComposeTransformsCopiesVariadicSlice(t *testing.T) {
	t.Parallel()
	addOne := function.Transform[int](func(value int) int { return value + 1 })
	double := function.Transform[int](func(value int) int { return value * 2 })
	addTen := function.Transform[int](func(value int) int { return value + 10 })
	steps := []function.Transform[int]{addTen, double, addOne}
	transform := function.ComposeTransforms(steps...)
	steps[0] = func(value int) int { return -value }
	steps[2] = nil
	// ((5+1)×2)+10 → 22; an aliased slice would panic on nil or negate.
	if got, want := transform(5), 22; got != want {
		t.Fatalf("ComposeTransforms after caller mutation = %d, want %d", got, want)
	}
}
