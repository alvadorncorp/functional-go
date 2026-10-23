package immutable_test

import (
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// sortedItems returns s.Items() sorted, since Set iteration/Items order is
// unspecified and tests must not depend on it.
func sortedItems[T int | string](s immutable.Set[T]) []T {
	items := s.Items()
	slices.Sort(items)
	return items
}

func TestSetOf(t *testing.T) {
	t.Parallel()

	s := immutable.SetOf(1, 2, 2, 3)
	if got, want := s.Len(), 3; got != want {
		t.Fatalf("Len() = %d, want %d (duplicates must collapse)", got, want)
	}
}

func TestCollectSet(t *testing.T) {
	t.Parallel()

	seq := slices.Values([]string{"a", "b", "a"})
	s := immutable.CollectSet(seq)

	if got, want := s.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if !s.Has("a") || !s.Has("b") {
		t.Fatalf("Has(a)=%v Has(b)=%v, want both true", s.Has("a"), s.Has("b"))
	}
}

func TestSetHas(t *testing.T) {
	t.Parallel()

	s := immutable.SetOf(1, 2, 3)

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

	base := immutable.SetOf(1, 2)
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

	base := immutable.SetOf(1, 2)
	result := base.Add(1)

	if !result.Equal(base) {
		t.Fatalf("Add(existing) = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestSetRemove(t *testing.T) {
	t.Parallel()

	base := immutable.SetOf(1, 2, 3)
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

	base := immutable.SetOf(1, 2)
	result := base.Remove(99)

	if !result.Equal(base) {
		t.Fatalf("Remove(absent) = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestSetValues(t *testing.T) {
	t.Parallel()

	s := immutable.SetOf(1, 2, 3)
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

	s := immutable.SetOf(1, 2, 3)
	items := s.Items()
	items[0] = 999

	if got, want := sortedItems(s), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Set changed after mutating Items() result: got %v, want %v", got, want)
	}
}

func TestSetIsEmpty(t *testing.T) {
	t.Parallel()

	var zero immutable.Set[int]
	if !zero.IsEmpty() {
		t.Fatal("zero Set.IsEmpty() = false, want true")
	}
	if !immutable.SetOf[int]().IsEmpty() {
		t.Fatal("empty SetOf.IsEmpty() = false, want true")
	}
	if immutable.SetOf(1).IsEmpty() {
		t.Fatal("non-empty Set.IsEmpty() = true, want false")
	}
}

func TestSetEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    immutable.Set[int]
		b    immutable.Set[int]
		want bool
	}{
		{"both empty", immutable.SetOf[int](), immutable.SetOf[int](), true},
		{"same elements different insertion order", immutable.SetOf(1, 2, 3), immutable.SetOf(3, 1, 2), true},
		{"different length", immutable.SetOf(1, 2), immutable.SetOf(1, 2, 3), false},
		{"different elements", immutable.SetOf(1, 2, 3), immutable.SetOf(1, 2, 4), false},
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

	v1 := immutable.SetOf(1)
	v2 := v1.Add(2, 3, 2)
	sibling := v1.Add(8, 9)
	v3 := v2.Add(3, 4, 5)
	built := v3.Builder().Add(6, 7).Remove(1, 2).Build()
	_ = built.Add(10, 11).Remove(3, 4)

	for i, version := range []struct {
		set  immutable.Set[int]
		want []int
	}{
		{v1, []int{1}},
		{v2, []int{1, 2, 3}},
		{sibling, []int{1, 8, 9}},
		{v3, []int{1, 2, 3, 4, 5}},
		{built, []int{3, 4, 5, 6, 7}},
	} {
		if got := sortedItems(version.set); !slices.Equal(got, version.want) {
			t.Fatalf("version %d = %v, want %v", i, got, version.want)
		}
	}
}

func TestSetRemovePreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := immutable.SetOf(1, 2, 3, 4, 5)
	v2 := v1.Remove(2, 4, 2, 99)
	sibling := v1.Remove(1, 3)
	v3 := v2.Remove(1, 5).Add(6, 7)
	built := v3.Builder().Remove(3, 6).Add(8, 9).Build()
	_ = built.Add(10, 11).Remove(7, 8)

	for i, version := range []struct {
		set  immutable.Set[int]
		want []int
	}{
		{v1, []int{1, 2, 3, 4, 5}},
		{v2, []int{1, 3, 5}},
		{sibling, []int{2, 4, 5}},
		{v3, []int{3, 6, 7}},
		{built, []int{7, 8, 9}},
	} {
		if got := sortedItems(version.set); !slices.Equal(got, version.want) {
			t.Fatalf("version %d = %v, want %v", i, got, version.want)
		}
	}
}

func TestSetZeroValueAdd(t *testing.T) {
	t.Parallel()

	var s immutable.Set[string]
	if got := s.Len(); got != 0 {
		t.Fatalf("zero Set Len() = %d, want 0", got)
	}
	if s.Has("a") {
		t.Fatal(`zero Set Has("a") = true, want false`)
	}

	s2 := s.Add("a")
	if !s2.Has("a") {
		t.Fatal(`Add from zero Has("a") = false, want true`)
	}
	if got := s2.Len(); got != 1 {
		t.Fatalf("Add from zero Len() = %d, want 1", got)
	}
	if got := s.Len(); got != 0 {
		t.Fatalf("zero Set mutated by Add: Len() = %d, want 0", got)
	}
	if s.Has("a") {
		t.Fatal(`zero Set Has("a") = true after Add on result, want false`)
	}
}

func TestSetZeroValueRemove(t *testing.T) {
	t.Parallel()

	var s immutable.Set[int]
	// Remove on a zero Set must materialize a seeded empty root rather than
	// freezing a zero seed into the lineage.
	removed := s.Remove(99)
	if removed.Len() != 0 {
		t.Fatalf("Remove from zero Len() = %d, want 0", removed.Len())
	}
	if s.Len() != 0 {
		t.Fatalf("zero Set mutated by Remove: Len() = %d, want 0", s.Len())
	}
	filled := removed.Add(1, 2)
	if got, want := sortedItems(filled), []int{1, 2}; !slices.Equal(got, want) {
		t.Fatalf("filled after Remove-from-zero = %v, want %v", got, want)
	}
}

func TestSetZeroValueBuilder(t *testing.T) {
	t.Parallel()

	var s immutable.Set[string]
	built := s.Builder().Add("a", "b").Build()

	if got, want := built.Len(), 2; got != want {
		t.Fatalf("Builder from zero Len() = %d, want %d", got, want)
	}
	if !built.Has("a") || !built.Has("b") {
		t.Fatalf("Builder from zero Has(a)=%v Has(b)=%v, want both true", built.Has("a"), built.Has("b"))
	}
	if s.Len() != 0 {
		t.Fatalf("zero Set mutated by Builder: Len() = %d, want 0", s.Len())
	}
}

func TestSetIndependentZeroValues(t *testing.T) {
	t.Parallel()

	var a, b immutable.Set[string]
	a1 := a.Add("x")
	b1 := b.Add("y")

	if !a1.Has("x") || a1.Has("y") {
		t.Fatalf("a1 Has(x)=%v Has(y)=%v, want true/false", a1.Has("x"), a1.Has("y"))
	}
	if !b1.Has("y") || b1.Has("x") {
		t.Fatalf("b1 Has(y)=%v Has(x)=%v, want true/false", b1.Has("y"), b1.Has("x"))
	}
	if a.Len() != 0 || b.Len() != 0 {
		t.Fatalf("zero receivers mutated: a.Len=%d b.Len=%d", a.Len(), b.Len())
	}
}
