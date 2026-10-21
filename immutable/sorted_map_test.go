package immutable_test

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

const (
	wantNilComparatorPanic = "immutable: nil comparator"
	wantZeroSortedMapPanic = "immutable: zero SortedMap has no comparator; " +
		"construct with EmptySortedMap or CollectSortedMap"
)

// reversed orders ints in descending order.
func reversed(a, b int) int { return cmp.Compare(b, a) }

// caseInsensitive treats keys that differ only in letter case as the same
// key, so which spelling survives a Set is observable through Keys.
func caseInsensitive(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

// assertSortedMap fails t unless m's keys and values, in iteration order,
// are exactly wantKeys and wantValues.
func assertSortedMap[K, V comparable](t *testing.T, m immutable.SortedMap[K, V], wantKeys []K, wantValues []V) {
	t.Helper()
	if got := slices.Collect(m.Keys()); !slices.Equal(got, wantKeys) {
		t.Fatalf("Keys() = %v, want %v", got, wantKeys)
	}
	if got := slices.Collect(m.Values()); !slices.Equal(got, wantValues) {
		t.Fatalf("Values() = %v, want %v", got, wantValues)
	}
	if got, want := m.Len(), len(wantKeys); got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
}

func TestSortedMapNilComparatorPanics(t *testing.T) {
	t.Parallel()

	expectPanic(t, wantNilComparatorPanic, func() { immutable.EmptySortedMap[int, int](nil) })
	expectPanic(t, wantNilComparatorPanic, func() {
		immutable.CollectSortedMap(nil, maps.All(map[int]int{1: 1}))
	})
}

func TestSortedMapZeroValue(t *testing.T) {
	t.Parallel()

	var zero immutable.SortedMap[int, string]
	if !zero.IsEmpty() || zero.Len() != 0 {
		t.Fatalf("zero SortedMap IsEmpty() = %v, Len() = %d, want true, 0", zero.IsEmpty(), zero.Len())
	}
	if got, ok := zero.Get(1); got != "" || ok {
		t.Fatalf(`zero Get(1) = (%q, %v), want ("", false)`, got, ok)
	}
	if zero.Has(1) {
		t.Fatal("zero Has(1) = true, want false")
	}
	for k, v := range zero.All() {
		t.Fatalf("zero All() yielded (%d, %q), want nothing", k, v)
	}
	if deleted := zero.Delete(1); !deleted.IsEmpty() {
		t.Fatalf("zero Delete(1) Len() = %d, want 0", deleted.Len())
	}

	expectPanic(t, wantZeroSortedMapPanic, func() { zero.Set(1, "a") })
	expectPanic(t, wantZeroSortedMapPanic, func() { zero.Builder() })
	expectPanic(t, wantZeroSortedMapPanic, func() { zero.Delete(1).Set(1, "a") })
}

func TestSortedMapIteratesInComparatorOrder(t *testing.T) {
	t.Parallel()

	source := map[int]string{5: "e", 1: "a", 4: "d", 2: "b", 3: "c"}

	asc := immutable.CollectSortedMap(cmp.Compare[int], maps.All(source))
	assertSortedMap(t, asc, []int{1, 2, 3, 4, 5}, []string{"a", "b", "c", "d", "e"})

	desc := immutable.CollectSortedMap(reversed, maps.All(source))
	assertSortedMap(t, desc, []int{5, 4, 3, 2, 1}, []string{"e", "d", "c", "b", "a"})

	var gotKeys []int
	for k, v := range asc.All() {
		if source[k] != v {
			t.Fatalf("All() yielded (%d, %q), want value %q", k, v, source[k])
		}
		gotKeys = append(gotKeys, k)
		if len(gotKeys) == 2 {
			break
		}
	}
	if want := []int{1, 2}; !slices.Equal(gotKeys, want) {
		t.Fatalf("All() with early break yielded keys %v, want %v", gotKeys, want)
	}
	for k := range desc.Keys() {
		if k != 5 {
			t.Fatalf("first Keys() of descending map = %d, want 5", k)
		}
		break
	}
	for v := range desc.Values() {
		if v != "e" {
			t.Fatalf("first Values() of descending map = %q, want %q", v, "e")
		}
		break
	}
}

func TestSortedMapGetHas(t *testing.T) {
	t.Parallel()

	m := immutable.EmptySortedMap[string, int](cmp.Compare[string]).Set("a", 1)

	if got, ok := m.Get("a"); got != 1 || !ok {
		t.Fatalf(`Get("a") = (%d, %v), want (1, true)`, got, ok)
	}
	if got, ok := m.Get("missing"); got != 0 || ok {
		t.Fatalf(`Get("missing") = (%d, %v), want (0, false)`, got, ok)
	}
	if !m.Has("a") || m.Has("missing") {
		t.Fatalf(`Has("a") = %v, Has("missing") = %v, want true, false`, m.Has("a"), m.Has("missing"))
	}
}

func TestSortedMapPreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := immutable.EmptySortedMap[string, int](cmp.Compare[string]).Set("b", 2).Set("a", 1)
	v2 := v1.Set("c", 3).Set("a", 99)
	v3 := v2.Delete("b").Delete("missing")

	assertSortedMap(t, v1, []string{"a", "b"}, []int{1, 2})
	assertSortedMap(t, v2, []string{"a", "b", "c"}, []int{99, 2, 3})
	assertSortedMap(t, v3, []string{"a", "c"}, []int{99, 3})
}

// TestSortedMapKeepsComparatorWhenEmptied proves an emptied SortedMap still
// accepts writes: Delete of the last key keeps the comparator.
func TestSortedMapKeepsComparatorWhenEmptied(t *testing.T) {
	t.Parallel()

	emptied := immutable.EmptySortedMap[int, string](reversed).Set(1, "a").Delete(1)
	if !emptied.IsEmpty() {
		t.Fatalf("Set then Delete Len() = %d, want 0", emptied.Len())
	}
	refilled := emptied.Set(2, "b").Set(3, "c")
	assertSortedMap(t, refilled, []int{3, 2}, []string{"c", "b"})
}

// TestSortedMapBuilderEmptiedThenBuilt proves a builder emptied by Delete
// builds a SortedMap that keeps the comparator and accepts writes.
func TestSortedMapBuilderEmptiedThenBuilt(t *testing.T) {
	t.Parallel()

	seed := immutable.EmptySortedMap[int, string](reversed).Set(1, "a").Set(2, "b")
	built := seed.Builder().Delete(1).Delete(2).Build()
	if !built.IsEmpty() {
		t.Fatalf("emptied builder Build() Len() = %d, want 0", built.Len())
	}

	assertSortedMap(t, built.Set(3, "c").Set(4, "d"), []int{4, 3}, []string{"d", "c"})
	assertSortedMap(t, built.Builder().Set(5, "e").Build(), []int{5}, []string{"e"})
	assertSortedMap(t, seed, []int{2, 1}, []string{"b", "a"})
}

// TestSortedMapLatestEquivalentKeyWins pins the key-replacement rule: on
// every write path, an equivalent stored key is replaced by the new key and
// its value by the new value.
func TestSortedMapLatestEquivalentKeyWins(t *testing.T) {
	t.Parallel()

	empty := immutable.EmptySortedMap[string, int](caseInsensitive)
	paths := map[string]immutable.SortedMap[string, int]{
		"Set":     empty.Set("Go", 1).Set("b", 0).Set("GO", 2),
		"Builder": empty.Builder().Set("Go", 1).Set("b", 0).Set("GO", 2).Build(),
		"Collect": immutable.CollectSortedMap(caseInsensitive, func(yield func(string, int) bool) {
			_ = yield("Go", 1) && yield("b", 0) && yield("GO", 2)
		}),
	}
	for name, m := range paths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertSortedMap(t, m, []string{"b", "GO"}, []int{0, 2})
			if got, ok := m.Get("go"); got != 2 || !ok {
				t.Fatalf(`Get("go") = (%d, %v), want (2, true)`, got, ok)
			}
		})
	}
}

func TestSortedMapEqual(t *testing.T) {
	t.Parallel()

	of := func(m map[string]int) immutable.SortedMap[string, int] {
		return immutable.CollectSortedMap(cmp.Compare[string], maps.All(m))
	}
	tests := []struct {
		name string
		a, b immutable.SortedMap[string, int]
		want bool
	}{
		{"both empty", of(nil), of(nil), true},
		{"zero and empty", immutable.SortedMap[string, int]{}, of(nil), true},
		{"same entries", of(map[string]int{"a": 1, "b": 2}), of(nil).Set("b", 2).Set("a", 1), true},
		{"different value", of(map[string]int{"a": 1}), of(map[string]int{"a": 2}), false},
		{"different key set", of(map[string]int{"a": 1}), of(map[string]int{"b": 1}), false},
		{"different length", of(map[string]int{"a": 1}), of(map[string]int{"a": 1, "b": 2}), false},
		{"zero and non-empty", immutable.SortedMap[string, int]{}, of(map[string]int{"a": 1}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("a.Equal(b) = %v, want %v", got, tt.want)
			}
			if got := tt.b.Equal(tt.a); got != tt.want {
				t.Errorf("b.Equal(a) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortedMapEqualFunc(t *testing.T) {
	t.Parallel()

	of := func(m map[string]int) immutable.SortedMap[string, int] {
		return immutable.CollectSortedMap(cmp.Compare[string], maps.All(m))
	}
	same := func(a, b int) bool { return a == b }
	alwaysFalse := func(a, b int) bool { return false }
	sameAbs := func(a, b int) bool { return a == b || a == -b }

	tests := []struct {
		name  string
		a, b  immutable.SortedMap[string, int]
		equal func(int, int) bool
		want  bool
	}{
		{"both empty", of(nil), of(nil), alwaysFalse, true},
		{"same entries", of(map[string]int{"a": 1, "b": 2}), of(map[string]int{"b": 2, "a": 1}), same, true},
		{"different value", of(map[string]int{"a": 1}), of(map[string]int{"a": 2}), same, false},
		{"different key set same length", of(map[string]int{"a": 1}), of(map[string]int{"b": 1}), same, false},
		{"different length", of(map[string]int{"a": 1}), of(map[string]int{"a": 1, "b": 2}), same, false},
		{"equal rejects value", of(map[string]int{"a": 1}), of(map[string]int{"a": 1}), alwaysFalse, false},
		{"custom abs equality", of(map[string]int{"a": 1, "b": -2}), of(map[string]int{"a": -1, "b": 2}), sameAbs, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.EqualFunc(tt.b, tt.equal); got != tt.want {
				t.Errorf("EqualFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortedMapBuilderMethodsPanicAfterBuild(t *testing.T) {
	t.Parallel()

	// newBuilder returns the already-used-up builder together with the
	// SortedMap Build() already returned, so a subtest can assert the
	// panicking post-Build call left that SortedMap untouched.
	newBuilder := func() (*immutable.SortedMapBuilder[string, int], immutable.SortedMap[string, int]) {
		b := immutable.EmptySortedMap[string, int](cmp.Compare[string]).Set("a", 1).Builder()
		built := b.Build()
		return b, built
	}

	calls := map[string]func(b *immutable.SortedMapBuilder[string, int]){
		"Set":         func(b *immutable.SortedMapBuilder[string, int]) { b.Set("b", 2) },
		"Delete":      func(b *immutable.SortedMapBuilder[string, int]) { b.Delete("a") },
		"Len":         func(b *immutable.SortedMapBuilder[string, int]) { b.Len() },
		"Get":         func(b *immutable.SortedMapBuilder[string, int]) { b.Get("a") },
		"Has":         func(b *immutable.SortedMapBuilder[string, int]) { b.Has("a") },
		"Build again": func(b *immutable.SortedMapBuilder[string, int]) { b.Build() },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			b, built := newBuilder()
			expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { call(b) })
			assertSortedMap(t, built, []string{"a"}, []int{1})
		})
	}
}

func TestSortedMapBuilderDoesNotAffectSeed(t *testing.T) {
	t.Parallel()

	seed := immutable.EmptySortedMap[string, int](cmp.Compare[string]).Set("a", 1).Set("b", 2)
	b := seed.Builder()
	b.Set("a", 99).Set("c", 3).Delete("b")

	if got, ok := b.Get("a"); got != 99 || !ok || b.Has("b") || b.Len() != 2 {
		t.Fatalf(`builder Get("a") = (%d, %v), Has("b") = %v, Len() = %d, want (99, true), false, 2`,
			got, ok, b.Has("b"), b.Len())
	}
	assertSortedMap(t, seed, []string{"a", "b"}, []int{1, 2})

	assertSortedMap(t, b.Build(), []string{"a", "c"}, []int{99, 3})
	assertSortedMap(t, seed, []string{"a", "b"}, []int{1, 2})
}

// TestSortedMapBuilderSurvivesComparatorPanic proves a comparator panic
// during a builder write propagates and leaves the builder's prior contents
// intact and usable.
func TestSortedMapBuilderSurvivesComparatorPanic(t *testing.T) {
	t.Parallel()

	const boom = "comparator rejected key"
	panicky := func(a, b string) int {
		if a == "boom" || b == "boom" {
			panic(boom)
		}
		return cmp.Compare(a, b)
	}
	seed := immutable.CollectSortedMap(panicky, maps.All(map[string]int{"a": 1, "c": 3, "e": 5}))
	b := seed.Builder().Set("b", 2).Set("d", 4)

	expectPanic(t, boom, func() { b.Set("boom", 0) })
	expectPanic(t, boom, func() { b.Delete("boom") })

	assertSortedMap(t, b.Build(), []string{"a", "b", "c", "d", "e"}, []int{1, 2, 3, 4, 5})
	assertSortedMap(t, seed, []string{"a", "c", "e"}, []int{1, 3, 5})
}

// TestSortedMapBuildDoesNotCopy proves Build freezes in O(1): building a
// large builder allocates nothing, so it cannot copy or rebuild the tree.
func TestSortedMapBuildDoesNotCopy(t *testing.T) {
	const runs, size = 10, 10_000

	builders := make([]*immutable.SortedMapBuilder[int, int], runs+1) // +1 for AllocsPerRun's warm-up call
	for i := range builders {
		builders[i] = immutable.EmptySortedMap[int, int](cmp.Compare[int]).Builder()
		for k := range size {
			builders[i].Set(k, k)
		}
	}
	next := 0
	if allocs := testing.AllocsPerRun(runs, func() {
		_ = builders[next].Build()
		next++
	}); allocs != 0 {
		t.Fatalf("Build() allocs = %v, want 0", allocs)
	}
}
