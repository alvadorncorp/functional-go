package eq_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// fixedEqual has its own Equal method that always returns a fixed result,
// deliberately contradicting == for the values used in the test below. Its
// presence proves the method is consulted before the interface-== fallback.
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
		if eq.Equal(1, 2) {
			t.Fatal("Equal(1, 2) = true, want false")
		}
		if !eq.Equal("a", "a") {
			t.Fatal(`Equal("a", "a") = false, want true`)
		}
	})

	t.Run("a custom Equal method takes priority over ==", func(t *testing.T) {
		t.Parallel()

		// Same field value, so == would report true; the type's own Equal
		// always reports false, so the method must win for the overall
		// result to be false.
		a := fixedEqual{n: 7}
		b := fixedEqual{n: 7}

		if eq.Equal(a, b) {
			t.Fatal("Equal(a, b) = true, want false (fixedEqual.Equal must take priority)")
		}
	})

	t.Run("non-comparable values without Equal return false", func(t *testing.T) {
		t.Parallel()

		// Slices are not comparable and have no Equal method: Equal must
		// not panic and must not deep-compare; both equal-looking and
		// unequal slices report false.
		if eq.Equal([]int{1}, []int{1}) {
			t.Fatal("Equal([]int{1}, []int{1}) = true, want false (no DeepEqual fallback)")
		}
		if eq.Equal([]int{1}, []int{2}) {
			t.Fatal("Equal([]int{1}, []int{2}) = true, want false")
		}
		if eq.Equal(map[string]int{"a": 1}, map[string]int{"a": 1}) {
			t.Fatal("Equal(map, map) = true, want false (no DeepEqual fallback)")
		}

		// Nil non-comparables: still no DeepEqual and no panic.
		var nilSlice, nilSlice2 []int
		if eq.Equal(nilSlice, nilSlice2) {
			t.Fatal("Equal(([]int)(nil), ([]int)(nil)) = true, want false")
		}
		var nilMap, nilMap2 map[string]int
		if eq.Equal(nilMap, nilMap2) {
			t.Fatal("Equal((map)(nil), (map)(nil)) = true, want false")
		}
		if eq.Equal[any]([]int(nil), []int(nil)) {
			t.Fatal("Equal[any]([]int(nil), []int(nil)) = true, want false")
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

	t.Run("typed nils of a comparable type are equal", func(t *testing.T) {
		t.Parallel()

		var p, q *int
		if !eq.Equal(p, q) {
			t.Fatal("Equal((*int)(nil), (*int)(nil)) = false, want true")
		}
	})
}

func TestEqualComparable(t *testing.T) {
	t.Parallel()

	if !eq.EqualComparable(1, 1) {
		t.Fatal("EqualComparable(1, 1) = false, want true")
	}
	if eq.EqualComparable(1, 2) {
		t.Fatal("EqualComparable(1, 2) = true, want false")
	}
	if !eq.EqualComparable("x", "x") {
		t.Fatal(`EqualComparable("x", "x") = false, want true`)
	}
}
