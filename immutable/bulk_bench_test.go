package immutable_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"go.alvadorn.org/functional/immutable"
)

var (
	bulkListSink  immutable.List[int]
	bulkMapSink   immutable.Map[int, int]
	bulkSetSink   immutable.Set[int]
	bulkEqualSink bool
)

func BenchmarkCollectionConstructors(b *testing.B) {
	for _, n := range []int{0, 1, 32, 1024, 10_000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			items := make([]int, n)
			entries := make(map[int]int, n)
			for i := range items {
				items[i] = i
				entries[i] = i
			}
			b.Run("ListOf", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkListSink = immutable.ListOf(items...)
				}
			})
			b.Run("MapOf", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkMapSink = immutable.MapOf(entries)
				}
			})
			b.Run("SetOf", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = immutable.SetOf(items...)
				}
			})
			b.Run("CollectList", func(b *testing.B) {
				seq := slices.Values(items)
				b.ReportAllocs()
				for b.Loop() {
					bulkListSink = immutable.CollectList(seq)
				}
			})
			b.Run("CollectMap", func(b *testing.B) {
				seq := maps.All(entries)
				b.ReportAllocs()
				for b.Loop() {
					bulkMapSink = immutable.CollectMap(seq)
				}
			})
			b.Run("CollectSet", func(b *testing.B) {
				seq := slices.Values(items)
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = immutable.CollectSet(seq)
				}
			})
			duplicates := make([]int, n)
			for i := range duplicates {
				duplicates[i] = i % 8
			}
			b.Run("SetOfDuplicates", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = immutable.SetOf(duplicates...)
				}
			})
			b.Run("CollectMapDuplicates", func(b *testing.B) {
				seq := func(yield func(int, int) bool) {
					for i, k := range duplicates {
						if !yield(k, i) {
							return
						}
					}
				}
				b.ReportAllocs()
				for b.Loop() {
					bulkMapSink = immutable.CollectMap(seq)
				}
			})
		})
	}
}

func BenchmarkCollectionBulkEdits(b *testing.B) {
	for _, n := range []int{0, 1, 32, 1024, 10_000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			items := make([]int, n)
			duplicates := make([]int, n)
			for i := range items {
				items[i] = i + 32
				duplicates[i] = i % 8
			}
			list := makeBenchList(32)
			set := makeBenchSet(32)
			removable := set.Add(items...)
			b.Run("AppendAcrossTail32", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkListSink = list.Append(items...)
				}
			})
			b.Run("Add", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = set.Add(items...)
				}
			})
			b.Run("AddExistingDuplicates", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = set.Add(duplicates...)
				}
			})
			b.Run("Remove", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = removable.Remove(items...)
				}
			})
			b.Run("RemoveDuplicates", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					bulkSetSink = set.Remove(duplicates...)
				}
			})
		})
	}
}

func BenchmarkCollectionEqualityAndTailDelete(b *testing.B) {
	for _, n := range []int{0, 1, 32, 1024, 10_000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			list, otherList := makeBenchList(n), makeBenchList(n)
			m, otherMap := makeBenchMap(n), makeBenchMap(n)
			sorted, otherSorted := makeBenchSortedMap(n), makeBenchSortedMap(n)
			sharedList, sharedMap := list, m
			for _, check := range []struct {
				name    string
				compare func() bool
			}{
				{"ListDistinct", func() bool { return list.Equal(otherList) }},
				{"ListShared", func() bool { return list.Equal(sharedList) }},
				{"MapDistinct", func() bool { return m.Equal(otherMap) }},
				{"MapShared", func() bool { return m.Equal(sharedMap) }},
				{"SortedMapDistinct", func() bool { return sorted.Equal(otherSorted) }},
			} {
				b.Run(check.name, func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						bulkEqualSink = check.compare()
					}
				})
			}
			if n > 0 {
				b.Run("DeleteLast", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						bulkListSink = list.Delete(n - 1)
					}
				})
			}
		})
	}
}
