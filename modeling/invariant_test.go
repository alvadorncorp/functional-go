package modeling_test

import (
	"errors"
	"testing"

	"go.alvadorn.org/functional/modeling"
)

var (
	errInvariantA = errors.New("invariant A")
	errInvariantB = errors.New("invariant B")
)

func TestFailFastStopsAtFirstFailure(t *testing.T) {
	calls := 0
	combined := modeling.FailFast(
		modeling.Invariant[int](func(int) error { calls++; return errInvariantA }),
		modeling.Invariant[int](func(int) error { calls++; return errInvariantB }),
	)

	err := combined(1)
	if err != errInvariantA {
		t.Fatalf("error = %v, want identical errInvariantA", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestFailFastWithNoInvariantsSucceeds(t *testing.T) {
	if err := modeling.FailFast[int]()(1); err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
}

func TestAccumulateRunsAllAndJoinsFailures(t *testing.T) {
	calls := 0
	combined := modeling.Accumulate(
		modeling.Invariant[int](func(int) error { calls++; return errInvariantA }),
		modeling.Invariant[int](func(int) error { calls++; return nil }),
		modeling.Invariant[int](func(int) error { calls++; return errInvariantB }),
	)

	err := combined(1)
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
	if !errors.Is(err, errInvariantA) || !errors.Is(err, errInvariantB) {
		t.Fatalf("error = %v, want joined invariant errors", err)
	}
}

func TestAccumulateWithNoInvariantsSucceeds(t *testing.T) {
	if err := modeling.Accumulate[int]()(1); err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
}

func TestInvariantCompositionRejectsNil(t *testing.T) {
	for name, compose := range map[string]func(...modeling.Invariant[int]) modeling.Invariant[int]{
		"fail-fast":  modeling.FailFast[int],
		"accumulate": modeling.Accumulate[int],
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "modeling: nil invariant" {
					t.Fatalf("panic = %#v, want %q", got, "modeling: nil invariant")
				}
			}()
			compose(nil)
		})
	}
}

func TestFailFastCopiesVariadicSlice(t *testing.T) {
	invariants := []modeling.Invariant[int]{
		func(int) error { return nil },
		func(int) error { return errInvariantA },
	}
	combined := modeling.FailFast(invariants...)
	invariants[0], invariants[1] = nil, func(int) error { return errInvariantB }

	if err := combined(1); err != errInvariantA {
		t.Fatalf("error = %v, want identical errInvariantA", err)
	}
}

func TestAccumulateCopiesVariadicSlice(t *testing.T) {
	invariants := []modeling.Invariant[int]{
		func(int) error { return errInvariantA },
		func(int) error { return errInvariantB },
	}
	combined := modeling.Accumulate(invariants...)
	invariants[0], invariants[1] = nil, func(int) error { return nil }

	err := combined(1)
	if !errors.Is(err, errInvariantA) || !errors.Is(err, errInvariantB) {
		t.Fatalf("error = %v, want joined invariant errors", err)
	}
}
