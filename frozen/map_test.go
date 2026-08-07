package frozen_test

import (
	"maps"
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/frozen"
)

func TestMapAccessors(t *testing.T) {
	t.Parallel()

	data := map[string]int{"a": 1, "b": 2}

	tests := []struct {
		name      string
		m         frozen.Map[string, int]
		wantLen   int
		wantItems map[string]int
	}{
		{
			name:      "empty",
			m:         frozen.NewMap[string, int](nil),
			wantLen:   0,
			wantItems: nil,
		},
		{
			name:      "several",
			m:         frozen.NewMap(data),
			wantLen:   len(data),
			wantItems: data,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.m.Len(); got != tc.wantLen {
				t.Errorf("Len() = %d, want %d", got, tc.wantLen)
			}
			if got := tc.m.Items(); !maps.Equal(got, tc.wantItems) {
				t.Errorf("Items() = %v, want %v", got, tc.wantItems)
			}
			for k, want := range tc.wantItems {
				if got, ok := tc.m.Get(k); !ok || got != want {
					t.Errorf("Get(%q) = (%v, %v), want (%v, true)", k, got, ok, want)
				}
				if !tc.m.Has(k) {
					t.Errorf("Has(%q) = false, want true", k)
				}
			}
			if tc.m.Has("missing-key") {
				t.Error(`Has("missing-key") = true, want false`)
			}
			if got, ok := tc.m.Get("missing-key"); ok || got != 0 {
				t.Errorf(`Get("missing-key") = (%v, %v), want (0, false)`, got, ok)
			}
		})
	}
}

func TestMapAllKeysValues(t *testing.T) {
	t.Parallel()

	src := map[string]int{"a": 1, "b": 2, "c": 3}
	m := frozen.NewMap(src)

	gotAll := map[string]int{}
	for k, v := range m.All() {
		gotAll[k] = v
	}
	if !maps.Equal(gotAll, src) {
		t.Errorf("All() = %v, want %v", gotAll, src)
	}

	gotKeys := map[string]bool{}
	for k := range m.Keys() {
		gotKeys[k] = true
	}
	for k := range src {
		if !gotKeys[k] {
			t.Errorf("Keys() missing %q", k)
		}
	}

	gotValues := map[int]bool{}
	for v := range m.Values() {
		gotValues[v] = true
	}
	for _, v := range src {
		if !gotValues[v] {
			t.Errorf("Values() missing %v", v)
		}
	}
}

func TestMapEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    frozen.Map[string, int]
		b    frozen.Map[string, int]
		want bool
	}{
		{
			name: "equal regardless of key order",
			a:    frozen.NewMap(map[string]int{"a": 1, "b": 2}),
			b:    frozen.NewMap(map[string]int{"b": 2, "a": 1}),
			want: true,
		},
		{
			name: "different length",
			a:    frozen.NewMap(map[string]int{"a": 1}),
			b:    frozen.NewMap(map[string]int{"a": 1, "b": 2}),
			want: false,
		},
		{
			name: "different value",
			a:    frozen.NewMap(map[string]int{"a": 1}),
			b:    frozen.NewMap(map[string]int{"a": 2}),
			want: false,
		},
		{
			name: "both zero value",
			a:    frozen.Map[string, int]{},
			b:    frozen.Map[string, int]{},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.a.Equal(tc.b); got != tc.want {
				t.Errorf("Equal() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMapEqualFunc(t *testing.T) {
	t.Parallel()

	ints := frozen.NewMap(map[string]int{"a": 1, "b": 2})
	strs := frozen.NewMap(map[string]string{"a": "1", "b": "2"})
	sameText := func(x int, y string) bool { return strconv.Itoa(x) == y }

	if !frozen.MapEqualFunc(ints, strs, sameText) {
		t.Error("MapEqualFunc(ints, strs, sameText) = false, want true")
	}
	if frozen.MapEqualFunc(ints, frozen.NewMap(map[string]string{"a": "1"}), sameText) {
		t.Error("MapEqualFunc(...) = true for different lengths, want false")
	}
}

func TestMapCopiesOnConstruction(t *testing.T) {
	t.Parallel()

	src := map[string]int{"a": 1}
	m := frozen.NewMap(src)

	src["a"] = -1

	if got, _ := m.Get("a"); got != 1 {
		t.Errorf("mutating the source map leaked into Map: Get(\"a\") = %v, want 1", got)
	}
}

func TestMapCopiesOnItems(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(map[string]int{"a": 1})

	first := m.Items()
	first["a"] = -1
	second := m.Items()

	if second["a"] != 1 {
		t.Errorf("mutating an earlier Items() result leaked into a later call: got %v, want 1", second["a"])
	}
}

func TestMapZeroValue(t *testing.T) {
	t.Parallel()

	var m frozen.Map[string, int]

	if got := m.Len(); got != 0 {
		t.Errorf("zero Map Len() = %d, want 0", got)
	}
	for range m.All() {
		t.Fatal("zero Map All() iterated an element")
	}
	if got := m.Items(); got != nil {
		t.Errorf("zero Map Items() = %v, want nil", got)
	}
	if got, ok := m.Get("x"); ok || got != 0 {
		t.Errorf(`Get("x") = (%v, %v), want (0, false)`, got, ok)
	}
}

func TestCollectMap(t *testing.T) {
	t.Parallel()

	src := map[string]int{"a": 1, "b": 2}
	m := frozen.CollectMap(maps.All(src))

	if want := frozen.NewMap(src); !m.Equal(want) {
		t.Errorf("CollectMap(...) = %v, want %v", m.Items(), want.Items())
	}
}
