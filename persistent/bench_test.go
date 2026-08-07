package persistent_test

import (
	"fmt"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

// Benchmark sizes: small (fits easily in cache), medium, and large enough
// to make an O(n^2) loop-of-copies path visibly worse than linear.
const (
	benchSizeSmall  = 16
	benchSizeMedium = 1_024
	benchSizeLarge  = 65_536
)

var benchSizes = []int{benchSizeSmall, benchSizeMedium, benchSizeLarge}

func makeBenchList(n int) persistent.List[int] {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	return persistent.NewList(items...)
}

func makeBenchMap(n int) persistent.Map[int, int] {
	m := make(map[int]int, n)
	for i := range n {
		m[i] = i
	}
	return persistent.NewMap(m)
}

func makeBenchSet(n int) persistent.Set[int] {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	return persistent.NewSet(items...)
}

// BenchmarkListAppend measures a single Append call against a List already
// holding n elements: one full O(n) copy per call.
func BenchmarkListAppend(b *testing.B) {
	for _, n := range benchSizes {
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
// Append n times, each on the previous result. With today's copy-on-write
// engine this is the O(n^2) cliff: total work is 1+2+...+n copies, so
// walltime grows much faster than linearly with n. This number documents
// exactly why Builder exists.
func BenchmarkListAppendLoop(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				l := persistent.NewList[int]()
				for i := range n {
					l = l.Append(i)
				}
			}
		})
	}
}

// BenchmarkListBuilderAppendLoop is BenchmarkListAppendLoop's Builder
// counterpart: n Appends against one shared, mutable buffer. This is the
// correct path for building in a loop, and should scale roughly linearly
// with n where BenchmarkListAppendLoop scales quadratically.
func BenchmarkListBuilderAppendLoop(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				builder := persistent.NewList[int]().Builder()
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

// BenchmarkMapSet measures a single Set call against a Map already holding
// n entries: one full O(n) copy per call.
func BenchmarkMapSet(b *testing.B) {
	for _, n := range benchSizes {
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
// Builder: n Sets against one shared, mutable buffer.
func BenchmarkMapBuilderSet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				builder := persistent.NewMap[int, int](nil).Builder()
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
