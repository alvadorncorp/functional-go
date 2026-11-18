//go:build !race

package order_test

import (
	"testing"

	"go.alvadorn.org/functional/primitives/order"
)

var orderTestResult int

func TestDerivedOrdersDoNotAllocate(t *testing.T) {
	byAge := order.By(func(value person) int { return value.age }, order.Natural[int]())
	combined := order.Then(byAge, order.Reverse(byAge))
	a, b := person{age: 10}, person{age: 20}
	if allocs := testing.AllocsPerRun(1000, func() {
		orderTestResult = combined(a, b)
	}); allocs != 0 {
		t.Fatalf("comparison allocations = %v, want 0", allocs)
	}
}
