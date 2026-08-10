package order_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/order"
)

var benchmarkOrderResult int

func BenchmarkNatural(b *testing.B) {
	b.ReportAllocs()
	o := order.Natural[int]()
	for b.Loop() {
		benchmarkOrderResult = o(1, 2)
	}
}

func BenchmarkBy(b *testing.B) {
	b.ReportAllocs()
	o := order.By(func(v person) int { return v.age }, order.Natural[int]())
	a, c := person{age: 1}, person{age: 2}
	for b.Loop() {
		benchmarkOrderResult = o(a, c)
	}
}

func BenchmarkThenPrimaryHit(b *testing.B) {
	b.ReportAllocs()
	primary := order.By(func(v person) string { return v.last }, order.Natural[string]())
	secondary := order.By(func(v person) int { return v.age }, order.Natural[int]())
	o := order.Then(primary, secondary)
	a, c := person{last: "A"}, person{last: "B"}
	for b.Loop() {
		benchmarkOrderResult = o(a, c)
	}
}

func BenchmarkThenSecondaryHit(b *testing.B) {
	b.ReportAllocs()
	primary := order.By(func(v person) string { return v.last }, order.Natural[string]())
	secondary := order.By(func(v person) int { return v.age }, order.Natural[int]())
	o := order.Then(primary, secondary)
	a, c := person{last: "A", age: 1}, person{last: "A", age: 2}
	for b.Loop() {
		benchmarkOrderResult = o(a, c)
	}
}
