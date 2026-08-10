package immutable_test

import (
	"maps"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

func TestMapOf(t *testing.T) {
	t.Parallel()

	source := map[string]int{"a": 1, "b": 2}
	m := immutable.MapOf(source)

	if got, want := m.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	// The caller's map must not be retained: mutating the source map after
	// construction must not reach the Map.
	source["a"] = 999
	source["c"] = 3
	if got, ok := m.Get("a"); got != 1 || !ok {
		t.Fatalf("Get(%q) = (%d, %v) after external mutation of source, want (1, true)", "a", got, ok)
	}
	if m.Has("c") {
		t.Fatal(`Has("c") = true after external mutation of source, want false`)
	}
}

func TestCollectMap(t *testing.T) {
	t.Parallel()

	seq := maps.All(map[string]int{"x": 1, "y": 2})
	m := immutable.CollectMap(seq)

	if got, want := m.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if got, ok := m.Get("x"); got != 1 || !ok {
		t.Fatalf(`Get("x") = (%d, %v), want (1, true)`, got, ok)
	}
}

func TestMapGet(t *testing.T) {
	t.Parallel()

	m := immutable.MapOf(map[string]int{"a": 1})

	tests := []struct {
		name    string
		key     string
		want    int
		wantOK  bool
		explain string
	}{
		{"present key", "a", 1, true, "existing key returns its value"},
		{"absent key", "missing", 0, false, "missing key returns zero value and false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := m.Get(tt.key)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Get(%q) = (%d, %v), want (%d, %v): %s", tt.key, got, ok, tt.want, tt.wantOK, tt.explain)
			}
		})
	}
}

func TestMapHas(t *testing.T) {
	t.Parallel()

	m := immutable.MapOf(map[string]int{"a": 1})

	if !m.Has("a") {
		t.Error(`Has("a") = false, want true`)
	}
	if m.Has("missing") {
		t.Error(`Has("missing") = true, want false`)
	}
}

func TestMapSet(t *testing.T) {
	t.Parallel()

	base := immutable.MapOf(map[string]int{"a": 1})
	before := maps.Clone(base.Items())

	updated := base.Set("b", 2)

	if got, ok := updated.Get("b"); got != 2 || !ok {
		t.Fatalf(`Set result Get("b") = (%d, %v), want (2, true)`, got, ok)
	}
	if got, ok := updated.Get("a"); got != 1 || !ok {
		t.Fatalf(`Set result Get("a") = (%d, %v), want (1, true) (untouched key preserved)`, got, ok)
	}
	after := base.Items()
	if !maps.Equal(before, after) {
		t.Fatalf("receiver mutated by Set: before %v, after %v", before, after)
	}
}

func TestMapSetOverwritesExistingKey(t *testing.T) {
	t.Parallel()

	base := immutable.MapOf(map[string]int{"a": 1})
	updated := base.Set("a", 999)

	if got, ok := updated.Get("a"); got != 999 || !ok {
		t.Fatalf(`Get("a") = (%d, %v), want (999, true)`, got, ok)
	}
	if got, ok := base.Get("a"); got != 1 || !ok {
		t.Fatalf(`receiver Get("a") = (%d, %v), want (1, true) (unchanged)`, got, ok)
	}
}

func TestMapDelete(t *testing.T) {
	t.Parallel()

	base := immutable.MapOf(map[string]int{"a": 1, "b": 2})
	before := base.Items()

	deleted := base.Delete("a")

	if deleted.Has("a") {
		t.Fatal(`Delete result Has("a") = true, want false`)
	}
	if got, ok := deleted.Get("b"); got != 2 || !ok {
		t.Fatalf(`Delete result Get("b") = (%d, %v), want (2, true)`, got, ok)
	}
	after := base.Items()
	if !maps.Equal(before, after) {
		t.Fatalf("receiver mutated by Delete: before %v, after %v", before, after)
	}
}

func TestMapDeleteAbsentKeyIsNoOp(t *testing.T) {
	t.Parallel()

	base := immutable.MapOf(map[string]int{"a": 1})
	result := base.Delete("missing")

	if !result.Equal(base) {
		t.Fatalf("Delete(missing key) = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestMapAll(t *testing.T) {
	t.Parallel()

	source := map[string]int{"a": 1, "b": 2, "c": 3}
	m := immutable.MapOf(source)

	got := maps.Collect(m.All())
	if !maps.Equal(got, source) {
		t.Fatalf("All() collected = %v, want %v", got, source)
	}
}

func TestMapKeys(t *testing.T) {
	t.Parallel()

	m := immutable.MapOf(map[string]int{"a": 1, "b": 2})
	got := map[string]bool{}
	for k := range m.Keys() {
		got[k] = true
	}
	if want := (map[string]bool{"a": true, "b": true}); len(got) != len(want) || !got["a"] || !got["b"] {
		t.Fatalf("Keys() collected = %v, want keys a,b", got)
	}
}

func TestMapValues(t *testing.T) {
	t.Parallel()

	m := immutable.MapOf(map[string]int{"a": 1, "b": 2})
	sum := 0
	for v := range m.Values() {
		sum += v
	}
	if want := 3; sum != want {
		t.Fatalf("sum of Values() = %d, want %d", sum, want)
	}
}

func TestMapItemsCopies(t *testing.T) {
	t.Parallel()

	m := immutable.MapOf(map[string]int{"a": 1})
	items := m.Items()
	items["a"] = 999
	items["new"] = 1

	if got, _ := m.Get("a"); got != 1 {
		t.Fatalf(`Get("a") = %d after mutating Items() result, want 1 (Map unaffected)`, got)
	}
	if m.Has("new") {
		t.Fatal(`Has("new") = true after mutating Items() result, want false`)
	}
}

func TestMapItemsIsNeverNil(t *testing.T) {
	t.Parallel()

	var zero immutable.Map[string, int]
	items := zero.Items()
	// Must be safe to write into: a nil map here would panic on assignment.
	items["a"] = 1
	if got := items["a"]; got != 1 {
		t.Fatalf("write into Items() result of zero Map failed, got %d want 1", got)
	}
}

func TestMapEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    immutable.Map[string, int]
		b    immutable.Map[string, int]
		want bool
	}{
		{"both empty", immutable.MapOf(map[string]int{}), immutable.MapOf(map[string]int{}), true},
		{
			"same entries",
			immutable.MapOf(map[string]int{"a": 1, "b": 2}),
			immutable.MapOf(map[string]int{"b": 2, "a": 1}),
			true,
		},
		{
			"different value",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"a": 2}),
			false,
		},
		{
			"different key set",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"b": 1}),
			false,
		},
		{
			"different length",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"a": 1, "b": 2}),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapIsEmpty(t *testing.T) {
	t.Parallel()

	var zero immutable.Map[string, int]
	if !zero.IsEmpty() {
		t.Fatal("zero Map.IsEmpty() = false, want true")
	}
	if !immutable.MapOf(map[string]int{}).IsEmpty() {
		t.Fatal("empty MapOf.IsEmpty() = false, want true")
	}
	if immutable.MapOf(map[string]int{"a": 1}).IsEmpty() {
		t.Fatal("non-empty Map.IsEmpty() = true, want false")
	}
}

func TestMapEqualFunc(t *testing.T) {
	t.Parallel()

	eq := func(a, b int) bool { return a == b }
	alwaysFalse := func(a, b int) bool { return false }

	tests := []struct {
		name string
		a    immutable.Map[string, int]
		b    immutable.Map[string, int]
		eq   func(int, int) bool
		want bool
	}{
		{"both empty", immutable.MapOf(map[string]int{}), immutable.MapOf(map[string]int{}), eq, true},
		{
			"same entries",
			immutable.MapOf(map[string]int{"a": 1, "b": 2}),
			immutable.MapOf(map[string]int{"b": 2, "a": 1}),
			eq,
			true,
		},
		{
			"different value",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"a": 2}),
			eq,
			false,
		},
		{
			"different key set same length",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"b": 1}),
			eq,
			false,
		},
		{
			"different length",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"a": 1, "b": 2}),
			eq,
			false,
		},
		{
			"eq rejects value",
			immutable.MapOf(map[string]int{"a": 1}),
			immutable.MapOf(map[string]int{"a": 1}),
			alwaysFalse,
			false,
		},
		{
			"custom abs equality",
			immutable.MapOf(map[string]int{"a": 1, "b": -2}),
			immutable.MapOf(map[string]int{"a": -1, "b": 2}),
			func(a, b int) bool {
				if a < 0 {
					a = -a
				}
				if b < 0 {
					b = -b
				}
				return a == b
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.EqualFunc(tt.b, tt.eq); got != tt.want {
				t.Errorf("EqualFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapSetPreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := immutable.MapOf(map[string]int{"a": 1, "b": 2})
	v2 := v1.Set("c", 3)
	v3 := v2.Set("a", 99).Set("d", 4)

	if got, want := v1.Items(), map[string]int{"a": 1, "b": 2}; !maps.Equal(got, want) {
		t.Fatalf("v1.Items() = %v, want %v", got, want)
	}
	if got, want := v2.Items(), map[string]int{"a": 1, "b": 2, "c": 3}; !maps.Equal(got, want) {
		t.Fatalf("v2.Items() = %v, want %v", got, want)
	}
	if got, want := v3.Items(), map[string]int{"a": 99, "b": 2, "c": 3, "d": 4}; !maps.Equal(got, want) {
		t.Fatalf("v3.Items() = %v, want %v", got, want)
	}
}

func TestMapDeletePreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := immutable.MapOf(map[string]int{"a": 1, "b": 2, "c": 3})
	v2 := v1.Delete("b")
	v3 := v2.Delete("a").Set("d", 4)

	if got, want := v1.Items(), map[string]int{"a": 1, "b": 2, "c": 3}; !maps.Equal(got, want) {
		t.Fatalf("v1.Items() = %v, want %v", got, want)
	}
	if got, want := v2.Items(), map[string]int{"a": 1, "c": 3}; !maps.Equal(got, want) {
		t.Fatalf("v2.Items() = %v, want %v", got, want)
	}
	if got, want := v3.Items(), map[string]int{"c": 3, "d": 4}; !maps.Equal(got, want) {
		t.Fatalf("v3.Items() = %v, want %v", got, want)
	}
}

func TestMapZeroValueSet(t *testing.T) {
	t.Parallel()

	var m immutable.Map[string, int]
	if got := m.Len(); got != 0 {
		t.Fatalf("zero Map Len() = %d, want 0", got)
	}
	if m.Has("a") {
		t.Fatal(`zero Map Has("a") = true, want false`)
	}

	m2 := m.Set("a", 1)
	if got, ok := m2.Get("a"); got != 1 || !ok {
		t.Fatalf(`Set from zero Get("a") = (%d, %v), want (1, true)`, got, ok)
	}
	if got := m.Len(); got != 0 {
		t.Fatalf("zero Map mutated by Set: Len() = %d, want 0", got)
	}
	if m.Has("a") {
		t.Fatal(`zero Map Has("a") = true after Set on result, want false`)
	}
}

func TestMapZeroValueDelete(t *testing.T) {
	t.Parallel()

	var m immutable.Map[string, int]
	// Delete on a zero Map must materialize a seeded empty root rather than
	// freezing a zero seed into the lineage.
	deleted := m.Delete("missing")
	if deleted.Len() != 0 {
		t.Fatalf("Delete from zero Len() = %d, want 0", deleted.Len())
	}
	if m.Len() != 0 {
		t.Fatalf("zero Map mutated by Delete: Len() = %d, want 0", m.Len())
	}
	// Subsequent Set on the Delete result must still work.
	filled := deleted.Set("k", 7)
	if v, ok := filled.Get("k"); v != 7 || !ok {
		t.Fatalf(`filled Get("k") = (%d, %v), want (7, true)`, v, ok)
	}
}

func TestMapZeroValueBuilder(t *testing.T) {
	t.Parallel()

	var m immutable.Map[string, int]
	built := m.Builder().Set("a", 1).Set("b", 2).Build()

	if got, want := built.Len(), 2; got != want {
		t.Fatalf("Builder from zero Len() = %d, want %d", got, want)
	}
	if got, ok := built.Get("a"); got != 1 || !ok {
		t.Fatalf(`Get("a") = (%d, %v), want (1, true)`, got, ok)
	}
	if got, ok := built.Get("b"); got != 2 || !ok {
		t.Fatalf(`Get("b") = (%d, %v), want (2, true)`, got, ok)
	}
	if m.Len() != 0 {
		t.Fatalf("zero Map mutated by Builder: Len() = %d, want 0", m.Len())
	}
}

func TestMapIndependentZeroValues(t *testing.T) {
	t.Parallel()

	var a, b immutable.Map[string, int]
	a1 := a.Set("x", 1)
	b1 := b.Set("y", 2)

	if got, ok := a1.Get("x"); got != 1 || !ok {
		t.Fatalf(`a1.Get("x") = (%d, %v), want (1, true)`, got, ok)
	}
	if a1.Has("y") {
		t.Fatal(`a1.Has("y") = true, want false (independent zero maps)`)
	}
	if got, ok := b1.Get("y"); got != 2 || !ok {
		t.Fatalf(`b1.Get("y") = (%d, %v), want (2, true)`, got, ok)
	}
	if b1.Has("x") {
		t.Fatal(`b1.Has("x") = true, want false (independent zero maps)`)
	}
	if a.Len() != 0 || b.Len() != 0 {
		t.Fatalf("zero receivers mutated: a.Len=%d b.Len=%d", a.Len(), b.Len())
	}
}
