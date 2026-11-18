package eq_test

import (
	"testing"

	"go.alvadorn.org/functional/internal/eq"
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

type pointerEqual struct{ n int }

func (p *pointerEqual) Equal(other *pointerEqual) bool {
	if p == nil {
		return other != nil
	}
	return other != nil && p.n == other.n
}

type dynamicEqual struct{ match any }

func (p *dynamicEqual) Equal(other any) bool {
	if p == nil {
		return other == nil
	}
	return p.match == other
}

func TestForConcrete(t *testing.T) {
	t.Parallel()

	if !eq.For[int]()(7, 7) {
		t.Fatal("For[int] must retain comparable equality")
	}
	if eq.For[fixedEqual]()(fixedEqual{n: 7}, fixedEqual{n: 7}) {
		t.Fatal("For[fixedEqual] must prefer Equal over ==")
	}
	if eq.For[[]int]()([]int(nil), []int(nil)) {
		t.Fatal("For[[]int] must not deep-compare nil slices")
	}
	if !eq.For[*int]()(nil, nil) {
		t.Fatal("For[*int] must compare typed nil pointers")
	}
}

func TestForPointerEqualer(t *testing.T) {
	t.Parallel()

	equal := eq.For[*pointerEqual]()
	for _, tc := range []struct {
		name string
		a, b *pointerEqual
		want bool
	}{
		{"nil receiver method beats ==", nil, nil, false},
		{"nil receiver receives operand", nil, &pointerEqual{n: 1}, true},
		{"non-nil receiver is not the strategy probe", &pointerEqual{n: 1}, nil, false},
		{"distinct pointers use value method", &pointerEqual{n: 1}, &pointerEqual{n: 1}, true},
		{"different receiver values", &pointerEqual{n: 1}, &pointerEqual{n: 2}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := equal(tc.a, tc.b); got != tc.want {
				t.Fatalf("For[*pointerEqual]() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestForInterface(t *testing.T) {
	t.Parallel()

	equal := eq.For[any]()
	for _, tc := range []struct {
		name string
		a, b any
		want bool
	}{
		{"nil interfaces", nil, nil, true},
		{"nil versus typed nil", nil, (*int)(nil), false},
		{"typed nil pointers", (*int)(nil), (*int)(nil), true},
		{"different typed nils", (*int)(nil), (*string)(nil), false},
		{"mixed comparable types", 1, int64(1), false},
		{"comparable values", 7, 7, true},
		{"non-comparable values", []int{1}, []int{1}, false},
		{"typed nil map", map[int]int(nil), map[int]int(nil), false},
		{"dynamic Equal method", &dynamicEqual{match: "match"}, "match", true},
		{"receiver order matters", "match", &dynamicEqual{match: "match"}, false},
		{"typed nil Equal receiver", (*dynamicEqual)(nil), nil, true},
		{"nil has no Equal method", nil, (*dynamicEqual)(nil), false},
		{"Equal signature must match T", fixedEqual{n: 7}, fixedEqual{n: 7}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := equal(tc.a, tc.b); got != tc.want {
				t.Fatalf("For[any]() = %v, want %v", got, tc.want)
			}
		})
	}
}

type namedValue interface{ value() int }

type interfaceEqual int

func (v interfaceEqual) value() int { return int(v) }

func (interfaceEqual) Equal(namedValue) bool { return false }

func TestForNonEmptyInterface(t *testing.T) {
	t.Parallel()

	equal := eq.For[namedValue]()
	if !equal(nil, nil) {
		t.Fatal("For[namedValue] must retain nil interface equality")
	}
	if equal(nil, interfaceEqual(1)) {
		t.Fatal("For[namedValue] must distinguish nil from a value")
	}
	if equal(interfaceEqual(1), interfaceEqual(1)) {
		t.Fatal("For[namedValue] must resolve Equal on each dynamic value")
	}
}

type sliceEqual []int

func (s sliceEqual) Equal(other sliceEqual) bool { return len(s) == len(other) }

func TestForNilSliceEqualer(t *testing.T) {
	t.Parallel()

	equal := eq.For[sliceEqual]()
	if !equal(nil, nil) {
		t.Fatal("For[sliceEqual] must consult Equal even when the zero value is nil")
	}
	if equal(nil, sliceEqual{1}) {
		t.Fatal("For[sliceEqual] must forward the operands to Equal")
	}
}

type valueEqualer struct{}

func (valueEqualer) Equal(*valueEqualer) bool { return true }

func TestForNilValueReceiverPanicsAtComparison(t *testing.T) {
	t.Parallel()

	equal := eq.For[*valueEqualer]()
	defer func() {
		if recover() == nil {
			t.Fatal("For[*valueEqualer] must preserve the nil value-receiver panic")
		}
	}()
	equal(nil, nil)
}
