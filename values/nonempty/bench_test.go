package nonempty_test

import (
	"testing"

	"go.alvadorn.org/functional/values/nonempty"
)

var (
	benchmarkNonEmptyOfResult    nonempty.NonEmpty[int]
	benchmarkNonEmptyMapResult   nonempty.NonEmpty[int]
	benchmarkNonEmptyItemsResult []int
)

func BenchmarkNonEmptyOf(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkNonEmptyOfResult = nonempty.Of(1, 2, 3)
	}
}

func BenchmarkNonEmptyMap(b *testing.B) {
	b.ReportAllocs()
	n := nonempty.Of(1, 2, 3)
	for b.Loop() {
		benchmarkNonEmptyMapResult = nonempty.Map(n, func(v int) int { return v + 1 })
	}
}

func BenchmarkNonEmptyItems(b *testing.B) {
	b.ReportAllocs()
	n := nonempty.Of(1, 2, 3)
	for b.Loop() {
		benchmarkNonEmptyItemsResult = n.Items()
	}
}
