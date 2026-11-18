package specification_test

import (
	"fmt"
	"testing"

	"go.alvadorn.org/functional/modeling/specification"
)

var _ specification.Specification[int] = func(value int) bool { return value > 0 }

func assertSpecificationPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	fn()
}

func assertSpecificationPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestSpecificationHasNominalFunctionIdentity(t *testing.T) {
	isPositive := specification.Specification[int](func(value int) bool { return value > 0 })
	if !isPositive(1) || isPositive(0) {
		t.Fatal("Specification did not preserve the wrapped Boolean condition")
	}
	if got := fmt.Sprintf("%T", isPositive); got != "specification.Specification[int]" {
		t.Fatalf("type = %q, want specification.Specification[int]", got)
	}
}

func TestZeroSpecificationUsesOrdinaryNilFunctionBehavior(t *testing.T) {
	var zero specification.Specification[int]
	if zero != nil {
		t.Fatal("zero Specification must be nil")
	}
	assertSpecificationPanics(t, func() { zero(1) })
}
