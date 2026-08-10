package predicate_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/predicate"
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

func BenchmarkAll(b *testing.B) {
	b.ReportAllocs()
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	p := predicate.Predicate[int](func(v int) bool { return v > 0 })
	for b.Loop() {
		benchmarkPredicateResult = predicate.All(values, p)
	}
}

func BenchmarkAnyShortCircuit(b *testing.B) {
	b.ReportAllocs()
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	p := predicate.Predicate[int](func(v int) bool { return v == 1 })
	for b.Loop() {
		benchmarkPredicateResult = predicate.Any(values, p)
	}
}
