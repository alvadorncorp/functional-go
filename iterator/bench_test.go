package iterator_test

import (
	"fmt"
	"iter"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/iterator"
)

var iteratorBenchmarkSlice []int

func benchmarkValues(size int) []int {
	values := make([]int, size)
	for index := range values {
		values[index] = index
	}
	return values
}

func BenchmarkIteratorPipelines(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000, 100_000} {
		values := benchmarkValues(size)
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			benchmarkMapCollect(b, values)
			benchmarkFilterMapCollect(b, values)
			benchmarkFilterMapTake10(b, values)
			benchmarkFlatMapCollect(b, values)
		})
	}
}

func benchmarkMapCollect(b *testing.B, values []int) {
	b.Run("MapCollect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			seq := iterator.Map(iterator.FromSlice(values), func(v int) int { return v * 2 })
			iteratorBenchmarkSlice = iterator.Collect(seq)
		}
	})
	b.Run("MapCollectManual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := make([]int, len(values))
			for i, v := range values {
				out[i] = v * 2
			}
			iteratorBenchmarkSlice = out
		}
	})
}

func benchmarkFilterMapCollect(b *testing.B, values []int) {
	b.Run("FilterMapCollect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			filtered := iterator.Filter(iterator.FromSlice(values), func(v int) bool { return v%2 == 0 })
			mapped := iterator.Map(filtered, func(v int) int { return v * 2 })
			iteratorBenchmarkSlice = iterator.Collect(mapped)
		}
	})
	b.Run("FilterMapCollectManual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := make([]int, 0, len(values)/2)
			for _, v := range values {
				if v%2 == 0 {
					out = append(out, v*2)
				}
			}
			iteratorBenchmarkSlice = out
		}
	})
}

func benchmarkFilterMapTake10(b *testing.B, values []int) {
	b.Run("FilterMapTake10", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			filtered := iterator.Filter(iterator.FromSlice(values), func(v int) bool { return v%2 == 0 })
			mapped := iterator.Map(filtered, func(v int) int { return v * 2 })
			iteratorBenchmarkSlice = iterator.Collect(iterator.Take(mapped, 10))
		}
	})
	b.Run("FilterMapTake10Manual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := make([]int, 0, 10)
			for _, v := range values {
				if v%2 != 0 {
					continue
				}
				out = append(out, v*2)
				if len(out) == 10 {
					break
				}
			}
			iteratorBenchmarkSlice = out
		}
	})
}

func benchmarkFlatMapCollect(b *testing.B, values []int) {
	b.Run("FlatMapCollect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			seq := iterator.FlatMap(iterator.FromSlice(values), func(v int) iter.Seq[int] {
				return iterator.FromValues(v, -v)
			})
			iteratorBenchmarkSlice = iterator.Collect(seq)
		}
	})
	b.Run("FlatMapCollectManual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := make([]int, 0, len(values)*2)
			for _, v := range values {
				out = append(out, v, -v)
			}
			iteratorBenchmarkSlice = out
		}
	})
}
