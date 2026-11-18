package immutable_test

import (
	"slices"
	"testing"

	"go.alvadorn.org/functional/immutable"
)

// setAlgebraCase pairs two operands with the independently written sorted
// results of every set-algebra operation over them.
type setAlgebraCase struct {
	name   string
	a, b   immutable.Set[int]
	union  []int
	inter  []int
	diffAB []int
	diffBA []int
}

func setAlgebraCases() []setAlgebraCase {
	var zero immutable.Set[int]
	shared := immutable.SetOf(1, 2, 3, 4)
	return []setAlgebraCase{
		{"disjoint", immutable.SetOf(1, 2), immutable.SetOf(3, 4), []int{1, 2, 3, 4}, nil, []int{1, 2}, []int{3, 4}},
		{
			"overlapping", immutable.SetOf(1, 2, 3), immutable.SetOf(2, 3, 4),
			[]int{1, 2, 3, 4},
			[]int{2, 3},
			[]int{1},
			[]int{4},
		},
		{"subset", immutable.SetOf(1, 2), immutable.SetOf(1, 2, 3), []int{1, 2, 3}, []int{1, 2}, nil, []int{3}},
		{"superset", immutable.SetOf(1, 2, 3), immutable.SetOf(2), []int{1, 2, 3}, []int{2}, []int{1, 3}, nil},
		{"empty", immutable.SetOf(1, 2), immutable.SetOf[int](), []int{1, 2}, nil, []int{1, 2}, nil},
		{"zero Set", zero, immutable.SetOf(1, 2), []int{1, 2}, nil, nil, []int{1, 2}},
		{"both zero", zero, zero, nil, nil, nil, nil},
		{"same root", shared, shared, []int{1, 2, 3, 4}, []int{1, 2, 3, 4}, nil, nil},
		{"independent seed", immutable.SetOf(1, 2, 3), immutable.SetOf(3, 2, 1), []int{1, 2, 3}, []int{1, 2, 3}, nil, nil},
		{
			"same lineage", shared, shared.Remove(1).Add(5),
			[]int{1, 2, 3, 4, 5},
			[]int{2, 3, 4},
			[]int{1},
			[]int{5},
		},
	}
}

func TestSetAlgebraModel(t *testing.T) {
	t.Parallel()

	for _, tt := range setAlgebraCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wantA, wantB := sortedItems(tt.a), sortedItems(tt.b)

			results := []struct {
				op   string
				got  immutable.Set[int]
				want []int
			}{
				{"a.Union(b)", tt.a.Union(tt.b), tt.union},
				{"b.Union(a)", tt.b.Union(tt.a), tt.union},
				{"a.Intersection(b)", tt.a.Intersection(tt.b), tt.inter},
				{"b.Intersection(a)", tt.b.Intersection(tt.a), tt.inter},
				{"a.Difference(b)", tt.a.Difference(tt.b), tt.diffAB},
				{"b.Difference(a)", tt.b.Difference(tt.a), tt.diffBA},
			}
			for _, r := range results {
				if got := sortedItems(r.got); !slices.Equal(got, r.want) {
					t.Errorf("%s = %v, want %v", r.op, got, r.want)
				}
				// Deriving from a result that may share an operand's
				// structure must never reach back into that operand.
				_ = r.got.Add(100).Remove(1, 2)
				_ = r.got.Builder().Add(101).Remove(3, 4).Build()
			}
			if got := sortedItems(tt.a); !slices.Equal(got, wantA) {
				t.Errorf("operand a changed: got %v, want %v", got, wantA)
			}
			if got := sortedItems(tt.b); !slices.Equal(got, wantB) {
				t.Errorf("operand b changed: got %v, want %v", got, wantB)
			}
		})
	}
}

// setLawSamples spans the zero and empty Sets, both seed lineages, and a Set
// large enough to need more than one HAMT level.
func setLawSamples() []immutable.Set[int] {
	var zero immutable.Set[int]
	bigBuilder := immutable.SetOf[int]().Builder()
	for i := range 100 {
		bigBuilder.Add(i)
	}
	big := bigBuilder.Build()
	small := immutable.SetOf(1, 2, 3)
	return []immutable.Set[int]{
		zero,
		immutable.SetOf[int](),
		small,
		small.Remove(1).Add(4, 5),
		immutable.SetOf(3, 4, 5, 6),
		big,
		big.Remove(0, 50, 99).Add(200),
	}
}

func TestSetAlgebraLaws(t *testing.T) {
	t.Parallel()

	samples := setLawSamples()
	var empty immutable.Set[int]
	for i, a := range samples {
		checkSetUnaryLaws(t, i, a, empty)
		for j, b := range samples {
			checkSetBinaryLaws(t, i, j, a, b)
			for k, c := range samples {
				checkSetAssociativity(t, [3]int{i, j, k}, a, b, c)
			}
		}
	}
}

func checkSetUnaryLaws(t *testing.T, i int, a, empty immutable.Set[int]) {
	t.Helper()

	laws := []struct {
		name      string
		got, want immutable.Set[int]
	}{
		{"a∪a = a", a.Union(a), a},
		{"a∩a = a", a.Intersection(a), a},
		{"a∪∅ = a", a.Union(empty), a},
		{"∅∪a = a", empty.Union(a), a},
		{"a∩∅ = ∅", a.Intersection(empty), empty},
		{"∅∩a = ∅", empty.Intersection(a), empty},
		{"a\\a = ∅", a.Difference(a), empty},
		{"a\\a' = ∅ (independent copy)", a.Difference(immutable.SetOf(a.Items()...)), empty},
		{"a\\∅ = a", a.Difference(empty), a},
		{"∅\\a = ∅", empty.Difference(a), empty},
	}
	for _, law := range laws {
		if !law.got.Equal(law.want) {
			t.Errorf("sample %d: %s violated: got %v, want %v",
				i, law.name, sortedItems(law.got), sortedItems(law.want))
		}
	}
}

func checkSetBinaryLaws(t *testing.T, i, j int, a, b immutable.Set[int]) {
	t.Helper()

	laws := []struct {
		name      string
		got, want immutable.Set[int]
	}{
		{"a∪b = b∪a", a.Union(b), b.Union(a)},
		{"a∩b = b∩a", a.Intersection(b), b.Intersection(a)},
		{"(a\\b)∪(a∩b) = a", a.Difference(b).Union(a.Intersection(b)), a},
		{"(a\\b)∩b = ∅", a.Difference(b).Intersection(b), immutable.Set[int]{}},
	}
	for _, law := range laws {
		if !law.got.Equal(law.want) {
			t.Errorf("samples (%d,%d): %s violated: got %v, want %v",
				i, j, law.name, sortedItems(law.got), sortedItems(law.want))
		}
	}
	if got, want := a.Union(b).Len()+a.Intersection(b).Len(), a.Len()+b.Len(); got != want {
		t.Errorf("samples (%d,%d): |a∪b|+|a∩b| = %d, want |a|+|b| = %d", i, j, got, want)
	}
}

func checkSetAssociativity(t *testing.T, idx [3]int, a, b, c immutable.Set[int]) {
	t.Helper()

	if l, r := a.Union(b).Union(c), a.Union(b.Union(c)); !l.Equal(r) {
		t.Errorf("samples %v: (a∪b)∪c = %v, a∪(b∪c) = %v", idx, sortedItems(l), sortedItems(r))
	}
	if l, r := a.Intersection(b).Intersection(c), a.Intersection(b.Intersection(c)); !l.Equal(r) {
		t.Errorf("samples %v: (a∩b)∩c = %v, a∩(b∩c) = %v", idx, sortedItems(l), sortedItems(r))
	}
}
