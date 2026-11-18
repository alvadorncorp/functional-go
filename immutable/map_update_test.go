package immutable_test

import (
	"maps"
	"testing"

	"go.alvadorn.org/functional/immutable"
)

// updateCall records one invocation of an Update callback.
type updateCall struct {
	old     int
	present bool
}

// recordingUpdate returns an Update callback that appends its arguments to
// calls and returns next.
func recordingUpdate(calls *[]updateCall, next int) func(int, bool) int {
	return func(old int, present bool) int {
		*calls = append(*calls, updateCall{old, present})
		return next
	}
}

func TestMapUpdate(t *testing.T) {
	t.Parallel()

	var zero immutable.Map[string, int]
	base := immutable.MapOf(map[string]int{"a": 1, "b": 2})

	tests := []struct {
		name     string
		m        immutable.Map[string, int]
		key      string
		wantCall updateCall
		want     map[string]int
		wantBase map[string]int
	}{
		{
			name:     "present key",
			m:        base,
			key:      "a",
			wantCall: updateCall{old: 1, present: true},
			want:     map[string]int{"a": 10, "b": 2},
			wantBase: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "absent key",
			m:        base,
			key:      "c",
			wantCall: updateCall{old: 0, present: false},
			want:     map[string]int{"a": 1, "b": 2, "c": 10},
			wantBase: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "zero Map",
			m:        zero,
			key:      "a",
			wantCall: updateCall{old: 0, present: false},
			want:     map[string]int{"a": 10},
			wantBase: map[string]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var calls []updateCall
			got := tt.m.Update(tt.key, recordingUpdate(&calls, 10))

			if len(calls) != 1 || calls[0] != tt.wantCall {
				t.Fatalf("Update(%q) called fn with %v, want exactly [%v]", tt.key, calls, tt.wantCall)
			}
			if items := got.Items(); !maps.Equal(items, tt.want) {
				t.Fatalf("Update(%q).Items() = %v, want %v", tt.key, items, tt.want)
			}
			if items := tt.m.Items(); !maps.Equal(items, tt.wantBase) {
				t.Fatalf("receiver changed by Update(%q): Items() = %v, want %v", tt.key, items, tt.wantBase)
			}
		})
	}
}

func TestMapUpdateStoresZeroResult(t *testing.T) {
	t.Parallel()

	m := immutable.MapOf(map[string]int{"a": 1})
	got := m.Update("a", func(int, bool) int { return 0 })

	if v, ok := got.Get("a"); v != 0 || !ok {
		t.Fatalf(`Update returning zero: Get("a") = (%d, %v), want (0, true)`, v, ok)
	}
}

func TestMapUpdatePreservesOlderVersions(t *testing.T) {
	t.Parallel()

	inc := func(old int, _ bool) int { return old + 1 }
	v1 := immutable.MapOf(map[string]int{"a": 1, "b": 2})
	v2 := v1.Update("a", inc)
	v3 := v2.Update("a", inc).Update("c", inc)

	if got, want := v1.Items(), map[string]int{"a": 1, "b": 2}; !maps.Equal(got, want) {
		t.Fatalf("v1.Items() = %v, want %v", got, want)
	}
	if got, want := v2.Items(), map[string]int{"a": 2, "b": 2}; !maps.Equal(got, want) {
		t.Fatalf("v2.Items() = %v, want %v", got, want)
	}
	if got, want := v3.Items(), map[string]int{"a": 3, "b": 2, "c": 1}; !maps.Equal(got, want) {
		t.Fatalf("v3.Items() = %v, want %v", got, want)
	}
}

func TestMapZeroValueUpdate(t *testing.T) {
	t.Parallel()

	var m immutable.Map[string, int]
	// Update on a zero Map must materialize a seeded empty root, so the result
	// keeps working as an ordinary Map lineage.
	updated := m.Update("a", func(int, bool) int { return 1 })
	filled := updated.Set("b", 2).Delete("a")

	if got, want := filled.Items(), map[string]int{"b": 2}; !maps.Equal(got, want) {
		t.Fatalf("filled.Items() = %v, want %v", got, want)
	}
	if got, want := updated.Items(), map[string]int{"a": 1}; !maps.Equal(got, want) {
		t.Fatalf("updated.Items() = %v, want %v", got, want)
	}
	if m.Len() != 0 {
		t.Fatalf("zero Map mutated by Update: Len() = %d, want 0", m.Len())
	}
}
