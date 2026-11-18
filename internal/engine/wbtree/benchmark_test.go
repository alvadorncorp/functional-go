package wbtree

import (
	"cmp"
	"fmt"
	"math/bits"
	"testing"

	"go.alvadorn.org/functional/internal/core/ownership"
)

// Engine micro-benchmark sizes, matching hamt and vector so the engines
// compare at the same points.
var engineBenchSizes = []int{32, 1_024, 10_000}

// setAllocsPerLevel bounds the allocations of one pure Set per bit of n: a
// delta=3 tree is at most log_{4/3}(n+1) ≈ 2.41·log2(n+1) high, a Set copies
// one node per level of its path, and rotations add only a few more.
const setAllocsPerLevel = 3

func makeBenchRoot(n int) Root[int, int] {
	r := New[int, int](cmp.Compare[int])
	for i := range n {
		r = r.Set(i, i)
	}
	return r
}

// TestSetAllocsFollowDepth proves a single pure Set path-copies instead of
// copying the tree: its allocations grow with log2(n), not n.
func TestSetAllocsFollowDepth(t *testing.T) {
	for _, n := range engineBenchSizes {
		r := makeBenchRoot(n)
		limit := float64(setAllocsPerLevel * bits.Len(uint(n)))
		for name, key := range map[string]int{"new": n, "existing": n / 2} {
			allocs := testing.AllocsPerRun(100, func() { _ = r.Set(key, -1) })
			if allocs > limit {
				t.Errorf("n=%d %s: Set allocs = %v, want <= %v", n, name, allocs, limit)
			}
		}
	}
}

// insertAllocsPerKey pins the pure path: sequential Set calls in either key
// order force single rotations whose heavy child was just path-copied, and
// must not clone it again (that cost ~19.8 allocs per key at n=10000 against
// ~18.8 now).
const insertAllocsPerKey = 19.0

// TestSequentialInsertDoesNotRecloneRotatedChild is deliberately not parallel
// so AllocsPerRun stays deterministic.
func TestSequentialInsertDoesNotRecloneRotatedChild(t *testing.T) {
	const n = 10_000
	orders := map[string]func(i int) int{
		"ascending":  func(i int) int { return i },
		"descending": func(i int) int { return n - i },
	}
	for name, key := range orders {
		allocs := testing.AllocsPerRun(3, func() {
			r := New[int, int](cmp.Compare[int])
			for i := range n {
				r = r.Set(key(i), i)
			}
		})
		if perKey := allocs / n; perKey > insertAllocsPerKey {
			t.Errorf("%s Set allocs/key = %.2f, want <= %v", name, perKey, insertAllocsPerKey)
		}
	}
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
				r := New[int, int](cmp.Compare[int])
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
				r := New[int, int](cmp.Compare[int])
				for i := range n {
					r = r.Set(i, i)
				}
				_ = r
			}
		})
	}
}
