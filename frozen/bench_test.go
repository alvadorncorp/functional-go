package frozen_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/frozen"
)

// benchData backs both benchmarks below so they measure the same Slice.
var benchData = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// BenchmarkSliceValues and BenchmarkSliceItems are a deliberate pair: Values
// must report 0 allocs/op (it iterates the backing array in place) while
// Items reports 1 alloc/op (it always copies). Run with -benchmem to see the
// documented difference "prefer Values" is built on.
func BenchmarkSliceValues(b *testing.B) {
	b.ReportAllocs()

	s := frozen.NewSlice(benchData...)
	for b.Loop() {
		sum := 0
		for v := range s.Values() {
			sum += v
		}
		_ = sum
	}
}

func BenchmarkSliceItems(b *testing.B) {
	b.ReportAllocs()

	s := frozen.NewSlice(benchData...)
	for b.Loop() {
		_ = s.Items()
	}
}
