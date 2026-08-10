package fold_test

import (
	"fmt"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/fold"
	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

var (
	benchmarkFoldInt      int
	benchmarkFoldOptional optional.Optional[int]
	benchmarkFoldSlice    []int
	benchmarkFoldSizes    = []int{0, 1, 8, 32, 1_024, 65_536}
)

func valuesForBenchmark(size int) []int {
	values := make([]int, size)
	for index := range values {
		values[index] = index
	}
	return values
}

func BenchmarkFoldLeft(b *testing.B) {
	for _, size := range benchmarkFoldSizes {
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			values := valuesForBenchmark(size)
			for b.Loop() {
				benchmarkFoldInt = fold.FoldLeft(values, 0, add)
			}
		})
	}
}

func BenchmarkFoldRight(b *testing.B) {
	for _, size := range benchmarkFoldSizes {
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			values := valuesForBenchmark(size)
			for b.Loop() {
				benchmarkFoldInt = fold.FoldRight(values, 0, func(value, acc int) int {
					return value + acc
				})
			}
		})
	}
}

func BenchmarkReduce(b *testing.B) {
	for _, size := range benchmarkFoldSizes {
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			values := valuesForBenchmark(size)
			for b.Loop() {
				benchmarkFoldOptional = fold.Reduce(values, add)
			}
		})
	}
}

func BenchmarkScanLeft(b *testing.B) {
	for _, size := range benchmarkFoldSizes {
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			values := valuesForBenchmark(size)
			for b.Loop() {
				benchmarkFoldSlice = fold.ScanLeft(values, 0, add)
			}
		})
	}
}
