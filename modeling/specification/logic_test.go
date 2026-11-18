package specification_test

import (
	"strings"
	"testing"

	"go.alvadorn.org/functional/modeling/specification"
)

func TestAllAndAnyEmptyIdentities(t *testing.T) {
	if !specification.All[int]()(42) {
		t.Fatal("All() = false, want true")
	}
	if specification.Any[int]()(42) {
		t.Fatal("Any() = true, want false")
	}
}

func TestAllEvaluatesLeftToRightAndShortCircuits(t *testing.T) {
	var trace []string
	first := specification.Specification[int](func(int) bool {
		trace = append(trace, "first")
		return true
	})
	second := specification.Specification[int](func(int) bool {
		trace = append(trace, "second")
		return false
	})
	third := specification.Specification[int](func(int) bool {
		trace = append(trace, "third")
		return true
	})

	if specification.All(first, second, third)(0) {
		t.Fatal("All = true, want false")
	}
	if got := strings.Join(trace, ","); got != "first,second" {
		t.Fatalf("All order = %q, want first,second", got)
	}
}

func TestAnyEvaluatesLeftToRightAndShortCircuits(t *testing.T) {
	var trace []string
	first := specification.Specification[int](func(int) bool {
		trace = append(trace, "first")
		return false
	})
	second := specification.Specification[int](func(int) bool {
		trace = append(trace, "second")
		return true
	})
	third := specification.Specification[int](func(int) bool {
		trace = append(trace, "third")
		return false
	})

	if !specification.Any(first, second, third)(0) {
		t.Fatal("Any = false, want true")
	}
	if got := strings.Join(trace, ","); got != "first,second" {
		t.Fatalf("Any order = %q, want first,second", got)
	}
}

func TestAllCopiesVariadicSlice(t *testing.T) {
	alwaysTrue := specification.Specification[int](func(int) bool { return true })
	specifications := []specification.Specification[int]{alwaysTrue, alwaysTrue}
	all := specification.All(specifications...)
	specifications[0] = nil
	specifications[1] = func(int) bool { return false }

	if !all(0) {
		t.Fatal("All = false after caller mutation, want true")
	}
}

func TestAnyCopiesVariadicSlice(t *testing.T) {
	alwaysFalse := specification.Specification[int](func(int) bool { return false })
	specifications := []specification.Specification[int]{alwaysFalse, alwaysFalse}
	anySpec := specification.Any(specifications...)
	specifications[0] = nil
	specifications[1] = func(int) bool { return true }

	if anySpec(0) {
		t.Fatal("Any = true after caller mutation, want false")
	}
}

func TestNotNegatesSpecification(t *testing.T) {
	even := specification.Specification[int](func(value int) bool { return value%2 == 0 })
	odd := specification.Not(even)
	if !odd(3) || odd(4) {
		t.Fatal("Not did not negate Specification")
	}
}

func TestLogicRejectsNilSpecificationsEagerly(t *testing.T) {
	var nilSpecification specification.Specification[int]
	assertSpecificationPanic(t, "modeling/specification: nil specification", func() {
		specification.All(nilSpecification)
	})
	assertSpecificationPanic(t, "modeling/specification: nil specification", func() {
		specification.Any(nilSpecification)
	})
	assertSpecificationPanic(t, "modeling/specification: nil specification", func() {
		specification.Not(nilSpecification)
	})
}
