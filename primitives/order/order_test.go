package order_test

import (
	"testing"

	"go.alvadorn.org/functional/primitives/order"
)

type person struct {
	first string
	last  string
	age   int
}

func TestNaturalAndDerivedOperations(t *testing.T) {
	t.Parallel()

	ints := order.Natural[int]()
	if order.Compare(1, 2) >= 0 || ints(2, 1) <= 0 || ints(2, 2) != 0 {
		t.Fatal("natural comparison failed")
	}
	less := order.Less(ints)
	if !less(1, 2) || less(2, 1) || less(2, 2) {
		t.Fatal("Less boundaries failed")
	}
	greater := order.Greater(ints)
	if !greater(2, 1) || greater(1, 2) || greater(2, 2) {
		t.Fatal("Greater boundaries failed")
	}
	equal := order.Equal(ints)
	if !equal(2, 2) || equal(1, 2) || equal(2, 1) {
		t.Fatal("Equal boundaries failed")
	}
	if got := order.Min(ints, 2, 1); got != 1 {
		t.Fatalf("Min = %d", got)
	}
	if got := order.Max(ints, 2, 1); got != 2 {
		t.Fatalf("Max = %d", got)
	}
	first := person{first: "first", age: 20}
	second := person{first: "second", age: 20}
	byAge := order.By(func(value person) int { return value.age }, ints)
	if got := order.Min(byAge, first, second); got != first {
		t.Fatalf("tie Min = %+v, want first argument", got)
	}
	if got := order.Max(byAge, first, second); got != first {
		t.Fatalf("tie Max = %+v, want first argument", got)
	}
}

func TestReverseByAndThen(t *testing.T) {
	t.Parallel()

	ints := order.Natural[int]()
	if order.Reverse(ints)(1, 2) <= 0 {
		t.Fatal("Reverse failed")
	}
	byAge := order.By(func(value person) int { return value.age }, ints)
	if byAge(person{age: 20}, person{age: 30}) >= 0 {
		t.Fatal("By failed")
	}
	byLast := order.By(func(value person) string { return value.last }, order.Natural[string]())
	secondaryCalls := 0
	secondary := order.Order[person](func(a, b person) int { secondaryCalls++; return byAge(a, b) })
	combined := order.Then(byLast, secondary)
	if combined(person{last: "A", age: 40}, person{last: "B", age: 20}) >= 0 || secondaryCalls != 0 {
		t.Fatalf("Then primary hit calls = %d, want 0", secondaryCalls)
	}
	if combined(person{last: "A", age: 40}, person{last: "A", age: 20}) <= 0 || secondaryCalls != 1 {
		t.Fatalf("Then secondary result/calls failed: %d", secondaryCalls)
	}
}
