package persistent_test

import (
	"maps"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

func TestNewMap(t *testing.T) {
	t.Parallel()

	source := map[string]int{"a": 1, "b": 2}
	m := persistent.NewMap(source)

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
	m := persistent.CollectMap(seq)

	if got, want := m.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if got, ok := m.Get("x"); got != 1 || !ok {
		t.Fatalf(`Get("x") = (%d, %v), want (1, true)`, got, ok)
	}
}

func TestMapGet(t *testing.T) {
	t.Parallel()

	m := persistent.NewMap(map[string]int{"a": 1})

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

	m := persistent.NewMap(map[string]int{"a": 1})

	if !m.Has("a") {
		t.Error(`Has("a") = false, want true`)
	}
	if m.Has("missing") {
		t.Error(`Has("missing") = true, want false`)
	}
}

func TestMapSet(t *testing.T) {
	t.Parallel()

	base := persistent.NewMap(map[string]int{"a": 1})
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

	base := persistent.NewMap(map[string]int{"a": 1})
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

	base := persistent.NewMap(map[string]int{"a": 1, "b": 2})
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

	base := persistent.NewMap(map[string]int{"a": 1})
	result := base.Delete("missing")

	if !result.Equal(base) {
		t.Fatalf("Delete(missing key) = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestMapAll(t *testing.T) {
	t.Parallel()

	source := map[string]int{"a": 1, "b": 2, "c": 3}
	m := persistent.NewMap(source)

	got := map[string]int{}
	for k, v := range m.All() {
		got[k] = v
	}
	if !maps.Equal(got, source) {
		t.Fatalf("All() collected = %v, want %v", got, source)
	}
}

func TestMapKeys(t *testing.T) {
	t.Parallel()

	m := persistent.NewMap(map[string]int{"a": 1, "b": 2})
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

	m := persistent.NewMap(map[string]int{"a": 1, "b": 2})
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

	m := persistent.NewMap(map[string]int{"a": 1})
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

	var zero persistent.Map[string, int]
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
		a    persistent.Map[string, int]
		b    persistent.Map[string, int]
		want bool
	}{
		{"both empty", persistent.NewMap(map[string]int{}), persistent.NewMap(map[string]int{}), true},
		{
			"same entries",
			persistent.NewMap(map[string]int{"a": 1, "b": 2}),
			persistent.NewMap(map[string]int{"b": 2, "a": 1}),
			true,
		},
		{
			"different value",
			persistent.NewMap(map[string]int{"a": 1}),
			persistent.NewMap(map[string]int{"a": 2}),
			false,
		},
		{
			"different key set",
			persistent.NewMap(map[string]int{"a": 1}),
			persistent.NewMap(map[string]int{"b": 1}),
			false,
		},
		{
			"different length",
			persistent.NewMap(map[string]int{"a": 1}),
			persistent.NewMap(map[string]int{"a": 1, "b": 2}),
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

func TestMapSetPreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := persistent.NewMap(map[string]int{"a": 1})
	v2 := v1.Set("b", 2)
	v3 := v2.Set("a", 99)

	if got, ok := v1.Get("a"); got != 1 || !ok || v1.Has("b") {
		t.Fatalf("v1 changed: Get(a)=(%d,%v) Has(b)=%v", got, ok, v1.Has("b"))
	}
	if got, ok := v2.Get("a"); got != 1 || !ok {
		t.Fatalf("v2.Get(a) = (%d,%v), want (1,true)", got, ok)
	}
	if got, ok := v2.Get("b"); got != 2 || !ok {
		t.Fatalf("v2.Get(b) = (%d,%v), want (2,true)", got, ok)
	}
	if got, ok := v3.Get("a"); got != 99 || !ok {
		t.Fatalf("v3.Get(a) = (%d,%v), want (99,true)", got, ok)
	}
}
