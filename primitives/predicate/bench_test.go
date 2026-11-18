package predicate_test

import (
	"testing"

	"go.alvadorn.org/functional/primitives/predicate"
)

var benchmarkPredicateResult bool

func BenchmarkAnd1(b *testing.B) {
	b.ReportAllocs()
	p := predicate.And[int](func(v int) bool { return v > 0 })
	for b.Loop() {
		benchmarkPredicateResult = p(1)
	}
}

func BenchmarkAnd4(b *testing.B) {
	b.ReportAllocs()
	first := predicate.Predicate[int](func(v int) bool { return v > 0 })
	second := predicate.Predicate[int](func(v int) bool { return v > 0 })
	third := predicate.Predicate[int](func(v int) bool { return v > 0 })
	fourth := predicate.Predicate[int](func(v int) bool { return v > 0 })
	p := predicate.And(first, second, third, fourth)
	for b.Loop() {
		benchmarkPredicateResult = p(1)
	}
}

func BenchmarkOrShortCircuit(b *testing.B) {
	b.ReportAllocs()
	p := predicate.Or[int](func(int) bool { return true }, func(int) bool { return false })
	for b.Loop() {
		benchmarkPredicateResult = p(1)
	}
}
