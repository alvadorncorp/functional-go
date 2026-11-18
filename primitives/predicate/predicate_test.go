package predicate_test

import (
	"testing"

	"go.alvadorn.org/functional/primitives/predicate"
)

func TestLogicAndEmptyIdentities(t *testing.T) {
	t.Parallel()
	even := predicate.Predicate[int](func(value int) bool { return value%2 == 0 })
	positive := predicate.Predicate[int](func(value int) bool { return value > 0 })
	if !predicate.And[int]()(1) || predicate.Or[int]()(1) {
		t.Fatal("empty And must be true and empty Or must be false")
	}
	if !predicate.And(even, positive)(4) || predicate.And(even, positive)(-2) {
		t.Fatal("And returned an unexpected result")
	}
	if !predicate.Or(even, positive)(3) || predicate.Or(even, positive)(-3) {
		t.Fatal("Or returned an unexpected result")
	}
	if predicate.Not(even)(4) || !predicate.Not(even)(3) {
		t.Fatal("Not returned an unexpected result")
	}
}

func TestAndOrShortCircuitLeftToRight(t *testing.T) {
	t.Parallel()
	andCalls := 0
	and := predicate.And[int](
		func(int) bool { andCalls++; return false },
		func(int) bool { andCalls++; return true },
	)
	if and(1) || andCalls != 1 {
		t.Fatalf("And = false with %d calls, want 1", andCalls)
	}
	orCalls := 0
	or := predicate.Or[int](
		func(int) bool { orCalls++; return true },
		func(int) bool { orCalls++; return false },
	)
	if !or(1) || orCalls != 1 {
		t.Fatalf("Or = true with %d calls, want 1", orCalls)
	}
}

func TestComparablePredicates(t *testing.T) {
	t.Parallel()
	if !predicate.EqualTo("go")("go") || predicate.EqualTo("go")("rust") {
		t.Fatal("EqualTo failed")
	}
	if !predicate.NotEqualTo(3)(4) || predicate.NotEqualTo(3)(3) {
		t.Fatal("NotEqualTo failed")
	}
	oneOf := predicate.OneOf(1, 3, 5)
	if !oneOf(3) || oneOf(4) || predicate.OneOf[int]()(1) {
		t.Fatal("OneOf failed")
	}
}

func TestCombinatorsCopyVariadicSlice(t *testing.T) {
	t.Parallel()
	alwaysTrue := predicate.Predicate[int](func(int) bool { return true })
	alwaysFalse := predicate.Predicate[int](func(int) bool { return false })

	andArgs := []predicate.Predicate[int]{alwaysTrue}
	and := predicate.And(andArgs...)
	andArgs[0] = alwaysFalse
	if !and(1) {
		t.Fatal("And observed a later change to its argument slice")
	}

	orArgs := []predicate.Predicate[int]{alwaysFalse}
	or := predicate.Or(orArgs...)
	orArgs[0] = alwaysTrue
	if or(1) {
		t.Fatal("Or observed a later change to its argument slice")
	}

	values := []int{1, 3}
	oneOf := predicate.OneOf(values...)
	values[0] = 7
	if !oneOf(1) || oneOf(7) {
		t.Fatal("OneOf observed a later change to its argument slice")
	}
}
