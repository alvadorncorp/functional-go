package immutable_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"go.alvadorn.org/functional/immutable"
)

func TestBulkConstructors(t *testing.T) {
	t.Parallel()

	for _, n := range []int{0, 1, 32, 1024} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()
			items := make([]int, n)
			entries := make(map[int]int, n)
			for i := range items {
				items[i] = i
				entries[i] = i * 2
			}
			lists := []immutable.List[int]{immutable.ListOf(items...), immutable.CollectList(slices.Values(items))}
			ms := []immutable.Map[int, int]{immutable.MapOf(entries), immutable.CollectMap(maps.All(entries))}
			sets := []immutable.Set[int]{immutable.SetOf(items...), immutable.CollectSet(slices.Values(items))}
			if n > 0 {
				items[0], entries[0] = -1, -1
			}
			for _, list := range lists {
				_ = list.Append(-1, -2)
				if list.Len() != n {
					t.Fatalf("List Len = %d, want %d", list.Len(), n)
				}
				for i, v := range list.All() {
					if v != i {
						t.Fatalf("List At(%d) = %d, want %d", i, v, i)
					}
				}
			}
			for _, m := range ms {
				_ = m.Builder().Set(0, -1).Set(-1, -2).Build()
				if m.Len() != n {
					t.Fatalf("Map Len = %d, want %d", m.Len(), n)
				}
				for i := range n {
					if v, ok := m.Get(i); !ok || v != i*2 {
						t.Fatalf("Map Get(%d) = (%d, %v), want (%d, true)", i, v, ok, i*2)
					}
				}
			}
			for _, set := range sets {
				_ = set.Add(-1, -2).Remove(0, 1)
				if set.Len() != n {
					t.Fatalf("Set Len = %d, want %d", set.Len(), n)
				}
				for i := range n {
					if !set.Has(i) {
						t.Fatalf("Set missing %d", i)
					}
				}
			}
		})
	}
}
