package order_test

import (
	"math"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/order"
)

func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func TestNaturalOrderLawsAndReverseInvolution(t *testing.T) {
	t.Parallel()

	o := order.Natural[int]()
	values := []int{-2, -1, 0, 1, 2}
	for _, a := range values {
		if o(a, a) != 0 {
			t.Fatalf("reflexivity failed at %d", a)
		}
		for _, b := range values {
			if sign(o(a, b)) != -sign(o(b, a)) {
				t.Fatalf("antisymmetry failed for %d,%d", a, b)
			}
			if order.Reverse(order.Reverse(o))(a, b) != o(a, b) {
				t.Fatal("Reverse involution failed")
			}
			for _, c := range values {
				if o(a, b) <= 0 && o(b, c) <= 0 && o(a, c) > 0 {
					t.Fatal("transitivity failed")
				}
			}
		}
	}
}

func TestReverseByAndThenLaws(t *testing.T) {
	t.Parallel()

	extreme := order.Order[int](func(a, b int) int {
		if a < b {
			return math.MinInt
		}
		if a > b {
			return math.MaxInt
		}
		return 0
	})
	if got, want := order.Reverse(extreme)(1, 2), extreme(2, 1); got != want || got != math.MaxInt {
		t.Fatalf("Reverse extreme result = %d, want %d", got, want)
	}
	project := func(value person) int { return value.age }
	byAge := order.By(project, extreme)
	a := person{first: "A", age: 20}
	b := person{first: "B", age: 20}
	c := person{first: "C", age: 30}
	if got, want := byAge(a, c), extreme(project(a), project(c)); got != want {
		t.Fatalf("By projection result = %d, want %d", got, want)
	}
	if byAge(a, b) != 0 || sign(byAge(a, c)) != sign(byAge(b, c)) || sign(byAge(c, a)) != sign(byAge(c, b)) {
		t.Fatal("By equivalence consistency failed")
	}

	primary := order.Order[int](func(a, b int) int {
		if a/10 < b/10 {
			return -7
		}
		if a/10 > b/10 {
			return 9
		}
		return 0
	})
	secondary := order.Order[int](func(a, b int) int { return b - a })
	if got := order.Then(primary, secondary)(1, 20); got != -7 {
		t.Fatalf("Then primary result = %d, want -7", got)
	}
	if got, want := order.Then(primary, secondary)(2, 3), secondary(2, 3); got != want {
		t.Fatalf("Then secondary result = %d, want %d", got, want)
	}
}
