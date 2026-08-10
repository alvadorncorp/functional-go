package hamt

import (
	"fmt"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// Engine micro-benchmark sizes. Focus is path-copy / depth cost, not product
// suite coverage (see immutable/bench_test.go for that).
var engineBenchSizes = []int{32, 1_024, 10_000}

func makeBenchRoot(n int) Root[int, int] {
	r := Empty[int, int]()
	for i := range n {
		r = r.Set(i, i)
	}
	return r
}

// BenchmarkGet measures Get on a map of size n (mid-key hit).
func BenchmarkGet(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchRoot(n)
			key := n / 2
			for b.Loop() {
				_, _ = r.Get(key)
			}
		})
	}
}

// BenchmarkSetExisting measures Set of a key already present (update path).
func BenchmarkSetExisting(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchRoot(n)
			key := n / 2
			for b.Loop() {
				_ = r.Set(key, -1)
			}
		})
	}
}

// BenchmarkSetNew measures Set of a key outside the existing set (insert path).
func BenchmarkSetNew(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchRoot(n)
			key := n // never present in [0, n)
			for b.Loop() {
				_ = r.Set(key, key)
			}
		})
	}
}

// BenchmarkDelete measures Delete of a present key.
func BenchmarkDelete(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchRoot(n)
			key := n / 2
			for b.Loop() {
				_ = r.Delete(key)
			}
		})
	}
}

// BenchmarkTransientBatch measures n TransientSet updates under one owner,
// starting from empty. Contrasts with BenchmarkImmutableBatch.
func BenchmarkTransientBatch(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				owner := ownership.New()
				r := Empty[int, int]()
				for i := range n {
					r = r.TransientSet(owner, i, i)
				}
				owner.Invalidate()
				_ = r
			}
		})
	}
}

// BenchmarkImmutableBatch measures n immutable Set calls in sequence,
// starting from empty. Contrasts with BenchmarkTransientBatch.
func BenchmarkImmutableBatch(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				r := Empty[int, int]()
				for i := range n {
					r = r.Set(i, i)
				}
				_ = r
			}
		})
	}
}
