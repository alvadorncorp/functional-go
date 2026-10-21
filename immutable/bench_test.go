package immutable_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// Benchmark sizes for Baseline B (path-copy engines).
//
// Single-op Map.Set / Set.Add / List.Append should show allocs/op tracking a
// depth curve (roughly O(log32 n) nodes), NOT scaling linearly with n.
// ReportAllocs is set on every sub-benchmark so that property is visible in
// go test -bench output. Loop-of-immutable-update still allocates n roots;
// Builder is the bulk path. Native mutable baselines quantify immutability
// overhead.
const (
	benchSizeSmall = 16
	benchSize1K    = 1_024
	benchSize16K   = 16_384
	benchSize64K   = 65_536
	benchSize1M    = 1_000_000
)

// Single-op structure-sharing sizes: include 1M (setup cost only; -bench is
// not run by default unit tests).
var benchSizesSingleOp = []int{benchSize1K, benchSize16K, benchSize64K, benchSize1M}

// Loop / iteration / lookup sizes: omit 1M so loop benches stay practical.
var benchSizes = []int{benchSizeSmall, benchSize1K, benchSize16K, benchSize64K}

func makeBenchList(n int) immutable.List[int] {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	return immutable.ListOf(items...)
}

func makeBenchMap(n int) immutable.Map[int, int] {
	m := make(map[int]int, n)
	for i := range n {
		m[i] = i
	}
	return immutable.MapOf(m)
}

func makeBenchSortedMap(n int) immutable.SortedMap[int, int] {
	b := immutable.EmptySortedMap[int, int](cmp.Compare[int]).Builder()
	for i := range n {
		b.Set(i, i)
	}
	return b.Build()
}

func makeBenchSet(n int) immutable.Set[int] {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	return immutable.SetOf(items...)
}

// BenchmarkListAppend measures a single Append against a List already holding
// n elements. Watch allocs/op: path-copy vector depth curve, not O(n) full
// copy. Still creates one new root per call.
func BenchmarkListAppend(b *testing.B) {
	for _, n := range benchSizesSingleOp {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			base := makeBenchList(n)
			for b.Loop() {
				_ = base.Append(0)
			}
		})
	}
}

// BenchmarkListAppendLoop measures building a List of size n by calling
// Append n times, each on the previous result. Even with path-copy, this
// allocates n intermediate Lists — the reason Builder exists for bulk work.
func BenchmarkListAppendLoop(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				l := immutable.ListOf[int]()
				for i := range n {
					l = l.Append(i)
				}
			}
		})
	}
}

// BenchmarkListBuilderAppendLoop is BenchmarkListAppendLoop's Builder
// counterpart: n Appends under exclusive ownership, then O(1) Freeze on
// Build. Prefer this path for loop construction.
func BenchmarkListBuilderAppendLoop(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				builder := immutable.ListOf[int]().Builder()
				for i := range n {
					builder.Append(i)
				}
				_ = builder.Build()
			}
		})
	}
}

// BenchmarkListAt measures indexed read cost on a List of size n.
func BenchmarkListAt(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			l := makeBenchList(n)
			mid := n / 2
			for b.Loop() {
				_ = l.At(mid)
			}
		})
	}
}

// BenchmarkListValues measures a full iteration over a List of size n.
func BenchmarkListValues(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			l := makeBenchList(n)
			for b.Loop() {
				for v := range l.Values() {
					_ = v
				}
			}
		})
	}
}

// BenchmarkMapGet measures key lookup cost on a Map of size n.
func BenchmarkMapGet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			m := makeBenchMap(n)
			key := n / 2
			for b.Loop() {
				_, _ = m.Get(key)
			}
		})
	}
}

// BenchmarkMapSet measures a single Set against a Map already holding n
// entries. Watch allocs/op: HAMT path-copy depth curve — must not grow
// linearly with n (contrast historical full-copy COW).
func BenchmarkMapSet(b *testing.B) {
	for _, n := range benchSizesSingleOp {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			m := makeBenchMap(n)
			for b.Loop() {
				_ = m.Set(n, n)
			}
		})
	}
}

// BenchmarkMapUpdate measures a single Update of a present key against a Map
// of size n: one Get plus one Set, so allocs/op follows BenchmarkMapSet's
// depth curve, not n.
func BenchmarkMapUpdate(b *testing.B) {
	inc := func(old int, _ bool) int { return old + 1 }
	for _, n := range benchSizesSingleOp {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			m := makeBenchMap(n)
			key := n / 2
			for b.Loop() {
				_ = m.Update(key, inc)
			}
		})
	}
}

// BenchmarkMapBuilderSet is BenchmarkMapSet's loop-building counterpart via
// Builder: n Sets under exclusive ownership, then O(1) Freeze.
func BenchmarkMapBuilderSet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				builder := immutable.MapOf[int, int](nil).Builder()
				for i := range n {
					builder.Set(i, i)
				}
				_ = builder.Build()
			}
		})
	}
}

// BenchmarkSortedMapGet is BenchmarkMapGet's SortedMap counterpart: key
// lookup on a SortedMap of size n, comparable against Map at the same n.
func BenchmarkSortedMapGet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			m := makeBenchSortedMap(n)
			key := n / 2
			for b.Loop() {
				_, _ = m.Get(key)
			}
		})
	}
}

// BenchmarkSortedMapSet measures a single Set against a SortedMap already
// holding n entries. Watch allocs/op: weight-balanced path-copy follows
// log2(n), not n.
func BenchmarkSortedMapSet(b *testing.B) {
	for _, n := range benchSizesSingleOp {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			m := makeBenchSortedMap(n)
			for b.Loop() {
				_ = m.Set(n, n)
			}
		})
	}
}

// BenchmarkSortedMapBuilderSet is BenchmarkSortedMapSet's loop-building
// counterpart via Builder: n Sets under exclusive ownership, then O(1)
// Freeze.
func BenchmarkSortedMapBuilderSet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				builder := immutable.EmptySortedMap[int, int](cmp.Compare[int]).Builder()
				for i := range n {
					builder.Set(i, i)
				}
				_ = builder.Build()
			}
		})
	}
}

// BenchmarkSetHas measures membership-test cost on a Set of size n.
func BenchmarkSetHas(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			s := makeBenchSet(n)
			v := n / 2
			for b.Loop() {
				_ = s.Has(v)
			}
		})
	}
}

// BenchmarkSetAdd measures a single Add against a Set already holding n
// members. Watch allocs/op: same HAMT depth-curve property as Map.Set —
// not linear in n.
func BenchmarkSetAdd(b *testing.B) {
	for _, n := range benchSizesSingleOp {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			s := makeBenchSet(n)
			for b.Loop() {
				_ = s.Add(n)
			}
		})
	}
}

// BenchmarkSetRemove measures a single Remove of a present member against a
// Set of size n. Same depth-curve watch as BenchmarkSetAdd.
func BenchmarkSetRemove(b *testing.B) {
	for _, n := range benchSizesSingleOp {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			s := makeBenchSet(n)
			// Remove a key that exists so the path does real work; result discarded.
			victim := n / 2
			for b.Loop() {
				_ = s.Remove(victim)
			}
		})
	}
}

// makeBenchSetRange returns a Set holding the n consecutive ints from start.
func makeBenchSetRange(start, n int) immutable.Set[int] {
	b := immutable.SetOf[int]().Builder()
	for i := range n {
		b.Add(start + i)
	}
	return b.Build()
}

// benchSetAlgebra runs op over a large Set of n elements and a small Set of m
// elements, for every n in benchSizes and m in {benchSizeSmall, n}. The two
// Sets hold independent seeds. With overlap, small holds the first m elements
// of large (small ⊆ large); without it, the two are disjoint.
func benchSetAlgebra(b *testing.B, overlap bool, op func(small, large immutable.Set[int]) immutable.Set[int]) {
	for _, n := range benchSizes {
		for _, m := range slices.Compact([]int{benchSizeSmall, n}) {
			b.Run(fmt.Sprintf("n=%d/m=%d", n, m), func(b *testing.B) {
				b.ReportAllocs()
				large := makeBenchSetRange(0, n)
				start := n
				if overlap {
					start = 0
				}
				small := makeBenchSetRange(start, m)
				for b.Loop() {
					_ = op(small, large)
				}
			})
		}
	}
}

// BenchmarkSetUnion measures Union. The nochange case (small ⊆ large) must
// report the same allocs/op for every n and m: it walks small, probes large,
// and returns large without opening a builder. The disjoint case opens one
// builder and adds all of small.
func BenchmarkSetUnion(b *testing.B) {
	union := func(small, large immutable.Set[int]) immutable.Set[int] { return small.Union(large) }
	b.Run("nochange", func(b *testing.B) { benchSetAlgebra(b, true, union) })
	b.Run("disjoint", func(b *testing.B) { benchSetAlgebra(b, false, union) })
}

// BenchmarkSetIntersection measures Intersection. The nochange case
// (small ⊆ large) must report the same allocs/op for every n and m: it walks
// small, probes large, and returns small without opening a builder. The
// disjoint case opens one builder and removes all of small.
func BenchmarkSetIntersection(b *testing.B) {
	inter := func(small, large immutable.Set[int]) immutable.Set[int] { return small.Intersection(large) }
	b.Run("nochange", func(b *testing.B) { benchSetAlgebra(b, true, inter) })
	b.Run("disjoint", func(b *testing.B) { benchSetAlgebra(b, false, inter) })
}

// BenchmarkSetDifference measures Difference with the small and with the
// large operand as receiver. The nochange cases (disjoint operands) must
// report the same allocs/op for every n and m: they walk the smaller operand,
// probe the larger, and return the receiver without opening a builder. The
// subset cases open one builder and remove the m shared elements.
func BenchmarkSetDifference(b *testing.B) {
	smallRecv := func(small, large immutable.Set[int]) immutable.Set[int] { return small.Difference(large) }
	largeRecv := func(small, large immutable.Set[int]) immutable.Set[int] { return large.Difference(small) }
	b.Run("nochange/receiver-small", func(b *testing.B) { benchSetAlgebra(b, false, smallRecv) })
	b.Run("nochange/receiver-large", func(b *testing.B) { benchSetAlgebra(b, false, largeRecv) })
	b.Run("subset/receiver-small", func(b *testing.B) { benchSetAlgebra(b, true, smallRecv) })
	b.Run("subset/receiver-large", func(b *testing.B) { benchSetAlgebra(b, true, largeRecv) })
}

// BenchmarkNativeSliceAppendLoop is the native, mutable-slice baseline for
// BenchmarkListAppendLoop / BenchmarkListBuilderAppendLoop: append to a
// plain []int n times, in place. It quantifies the cost immutability adds.
func BenchmarkNativeSliceAppendLoop(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var s []int
				for i := range n {
					s = append(s, i)
				}
				_ = s
			}
		})
	}
}

// BenchmarkNativeMapSet is the native, mutable-map baseline for
// BenchmarkMapSet / BenchmarkMapBuilderSet: assign into a plain map[int]int
// n times, in place.
func BenchmarkNativeMapSet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m := make(map[int]int, n)
				for i := range n {
					m[i] = i
				}
			}
		})
	}
}

// BenchmarkNativeMapGet is the native, mutable-map baseline for
// BenchmarkMapGet: lookup on a plain map[int]int of size n.
func BenchmarkNativeMapGet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			m := make(map[int]int, n)
			for i := range n {
				m[i] = i
			}
			key := n / 2
			for b.Loop() {
				_ = m[key]
			}
		})
	}
}
