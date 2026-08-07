package persistent_test

import (
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

// sortedItems returns s.Items() sorted, since Set iteration/Items order is
// unspecified and tests must not depend on it.
func sortedItems[T int | string](s persistent.Set[T]) []T {
	items := s.Items()
	slices.Sort(items)
	return items
}

func TestNewSet(t *testing.T) {
	t.Parallel()

	s := persistent.NewSet(1, 2, 2, 3)
	if got, want := s.Len(), 3; got != want {
		t.Fatalf("Len() = %d, want %d (duplicates must collapse)", got, want)
	}
}

func TestCollectSet(t *testing.T) {
	t.Parallel()

	seq := slices.Values([]string{"a", "b", "a"})
	s := persistent.CollectSet(seq)

	if got, want := s.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if !s.Has("a") || !s.Has("b") {
		t.Fatalf("Has(a)=%v Has(b)=%v, want both true", s.Has("a"), s.Has("b"))
	}
}

func TestSetHas(t *testing.T) {
	t.Parallel()

	s := persistent.NewSet(1, 2, 3)

	tests := []struct {
		name string
		v    int
		want bool
	}{
		{"present", 2, true},
		{"absent", 99, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := s.Has(tt.v); got != tt.want {
				t.Errorf("Has(%d) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestSetAdd(t *testing.T) {
	t.Parallel()

	base := persistent.NewSet(1, 2)
	before := sortedItems(base)

	added := base.Add(3, 4)

	if got, want := sortedItems(added), []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Add result = %v, want %v", got, want)
	}
	after := sortedItems(base)
	if !slices.Equal(before, after) {
		t.Fatalf("receiver mutated by Add: before %v, after %v", before, after)
	}
}

func TestSetAddExistingIsNoOp(t *testing.T) {
	t.Parallel()

	base := persistent.NewSet(1, 2)
	result := base.Add(1)

	if !result.Equal(base) {
		t.Fatalf("Add(existing) = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestSetRemove(t *testing.T) {
	t.Parallel()

	base := persistent.NewSet(1, 2, 3)
	before := sortedItems(base)

	removed := base.Remove(2)

	if got, want := sortedItems(removed), []int{1, 3}; !slices.Equal(got, want) {
		t.Fatalf("Remove result = %v, want %v", got, want)
	}
	after := sortedItems(base)
	if !slices.Equal(before, after) {
		t.Fatalf("receiver mutated by Remove: before %v, after %v", before, after)
	}
}

func TestSetRemoveAbsentIsNoOp(t *testing.T) {
	t.Parallel()

	base := persistent.NewSet(1, 2)
	result := base.Remove(99)

	if !result.Equal(base) {
		t.Fatalf("Remove(absent) = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestSetValues(t *testing.T) {
	t.Parallel()

	s := persistent.NewSet(1, 2, 3)
	var got []int
	for v := range s.Values() {
		got = append(got, v)
	}
	slices.Sort(got)
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Values() collected = %v, want %v", got, want)
	}
}

func TestSetItemsCopies(t *testing.T) {
	t.Parallel()

	s := persistent.NewSet(1, 2, 3)
	items := s.Items()
	items[0] = 999

	if got, want := sortedItems(s), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Set changed after mutating Items() result: got %v, want %v", got, want)
	}
}

func TestSetEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    persistent.Set[int]
		b    persistent.Set[int]
		want bool
	}{
		{"both empty", persistent.NewSet[int](), persistent.NewSet[int](), true},
		{"same elements different insertion order", persistent.NewSet(1, 2, 3), persistent.NewSet(3, 1, 2), true},
		{"different length", persistent.NewSet(1, 2), persistent.NewSet(1, 2, 3), false},
		{"different elements", persistent.NewSet(1, 2, 3), persistent.NewSet(1, 2, 4), false},
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

func TestSetAddPreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := persistent.NewSet(1)
	v2 := v1.Add(2)
	v3 := v2.Add(3)

	if got, want := sortedItems(v1), []int{1}; !slices.Equal(got, want) {
		t.Fatalf("v1 = %v, want %v", got, want)
	}
	if got, want := sortedItems(v2), []int{1, 2}; !slices.Equal(got, want) {
		t.Fatalf("v2 = %v, want %v", got, want)
	}
	if got, want := sortedItems(v3), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("v3 = %v, want %v", got, want)
	}
}
