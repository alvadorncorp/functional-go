package vector

import (
	"fmt"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// Engine micro-benchmark sizes. Focus is path-copy / depth cost, not product
// suite coverage (see persistent/bench_test.go for that).
var engineBenchSizes = []int{32, 1_024, 10_000}

func makeBenchVector(n int) Root[int] {
	r := Empty[int]()
	// Transient build for setup only — not measured.
	owner := ownership.New()
	for i := range n {
		r = r.TransientAppend(owner, i)
	}
	owner.Invalidate()
	return r
}

// BenchmarkAt measures indexed read cost on a vector of size n.
func BenchmarkAt(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchVector(n)
			mid := n / 2
			for b.Loop() {
				_ = r.At(mid)
			}
		})
	}
}

// BenchmarkSet measures a single mid-index Set against a vector of size n.
// Watch allocs/op: path-copy depth curve, not O(n) full copy.
func BenchmarkSet(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchVector(n)
			mid := n / 2
			for b.Loop() {
				_ = r.Set(mid, -1)
			}
		})
	}
}

// BenchmarkAppend measures a single Append against a vector already holding n
// elements.
func BenchmarkAppend(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchVector(n)
			for b.Loop() {
				_ = r.Append(0)
			}
		})
	}
}

// BenchmarkPop measures a single Pop against a vector of size n.
func BenchmarkPop(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchVector(n)
			for b.Loop() {
				_, _ = r.Pop()
			}
		})
	}
}

// BenchmarkDeleteMid measures Delete at the middle index.
// Documented O(n) trade-off: sequential rebuild, not path-aware shift.
func BenchmarkDeleteMid(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := makeBenchVector(n)
			mid := n / 2
			for b.Loop() {
				_ = r.Delete(mid)
			}
		})
	}
}

// BenchmarkTransientAppendBatch measures n TransientAppends under one owner,
// starting from empty. Contrasts with BenchmarkPersistentAppendBatch.
func BenchmarkTransientAppendBatch(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				owner := ownership.New()
				r := Empty[int]()
				for i := range n {
					r = r.TransientAppend(owner, i)
				}
				owner.Invalidate()
				_ = r
			}
		})
	}
}

// BenchmarkPersistentAppendBatch measures n immutable Appends in sequence,
// starting from empty. Contrasts with BenchmarkTransientAppendBatch.
func BenchmarkPersistentAppendBatch(b *testing.B) {
	for _, n := range engineBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				r := Empty[int]()
				for i := range n {
					r = r.Append(i)
				}
				_ = r
			}
		})
	}
}
