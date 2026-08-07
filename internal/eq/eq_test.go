package eq_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// fixedEqual has its own Equal method that always returns a fixed result,
// deliberately contradicting both == and reflect.DeepEqual for the values
// used in the test below. Its presence proves tier 1 (the method) is
// consulted before tier 2 (==) or tier 3 (reflect.DeepEqual).
type fixedEqual struct {
	n int
}

func (fixedEqual) Equal(fixedEqual) bool { return false }

func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("comparable values use ==", func(t *testing.T) {
		t.Parallel()

		if !eq.Equal(1, 1) {
			t.Fatal("Equal(1, 1) = false, want true")
		}
	})

	t.Run("a custom Equal method takes priority over == and DeepEqual", func(t *testing.T) {
		t.Parallel()

		// Same field value, so both == and reflect.DeepEqual would report
		// true; the type's own Equal always reports false, so tier 1 must
		// win for the overall result to be false.
		a := fixedEqual{n: 7}
		b := fixedEqual{n: 7}

		if eq.Equal(a, b) {
			t.Fatal("Equal(a, b) = true, want false (fixedEqual.Equal must take priority)")
		}
	})

	t.Run("non-comparable values fall back to reflect.DeepEqual", func(t *testing.T) {
		t.Parallel()

		if !eq.Equal([]int{1}, []int{1}) {
			t.Fatal("Equal([]int{1}, []int{1}) = false, want true")
		}
	})

	t.Run("two nil interfaces are equal", func(t *testing.T) {
		t.Parallel()

		if !eq.Equal[any](nil, nil) {
			t.Fatal("Equal[any](nil, nil) = false, want true")
		}
	})

	t.Run("a nil and a non-nil interface are not equal", func(t *testing.T) {
		t.Parallel()

		if eq.Equal[any](nil, 1) {
			t.Fatal("Equal[any](nil, 1) = true, want false")
		}
	})
}
