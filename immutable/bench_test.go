package immutable_test

import (
	"fmt"
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
