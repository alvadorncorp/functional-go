package predicate

import "testing"

// Compatibility tests stay in the defining package while the v1 names remain supported.
func TestQuantifiers(t *testing.T) {
	t.Parallel()
	positive := Predicate[int](func(value int) bool { return value > 0 })
	for _, values := range [][]int{nil, {}} {
		if !All(values, positive) || Any(values, positive) || !None(values, positive) {
			t.Fatal("empty quantifier identities failed")
		}
		if !All(values, nil) || Any(values, nil) || !None(values, nil) {
			t.Fatal("empty quantifiers must not call a nil predicate")
		}
	}
	values := []int{1, 2, 3}
	if !All(values, positive) || !Any(values, EqualTo(2)) || None(values, positive) {
		t.Fatal("quantifier result failed")
	}
	if All([]int{1, 0, 3}, positive) || Any([]int{-1, 0}, positive) || !None([]int{-1, 0}, positive) {
		t.Fatal("quantifier result with nonmatching values failed")
	}
}

func TestQuantifiersShortCircuitLeftToRight(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		quantifier func([]int, Predicate[int]) bool
		match      func(int) bool
		want       bool
	}{
		{name: "All", quantifier: All[int], match: func(value int) bool { return value < 2 }, want: false},
		{name: "Any", quantifier: Any[int], match: func(value int) bool { return value == 2 }, want: true},
		{name: "None", quantifier: None[int], match: func(value int) bool { return value == 2 }, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got := tc.quantifier([]int{1, 2, 3}, func(value int) bool {
				calls++
				if value != calls {
					t.Fatalf("callback value = %d, want %d", value, calls)
				}
				return tc.match(value)
			})
			if got != tc.want || calls != 2 {
				t.Fatalf("result = %t with %d calls, want %t with 2 calls", got, calls, tc.want)
			}
		})
	}
}

func TestQuantifiersNilPredicatePanicsForNonemptyValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		quantifier func([]int, Predicate[int]) bool
	}{
		{name: "All", quantifier: All[int]},
		{name: "Any", quantifier: Any[int]},
		{name: "None", quantifier: None[int]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("nonempty quantifier did not panic for a nil predicate")
				}
			}()
			tc.quantifier([]int{1}, nil)
		})
	}
}

func TestQuantifierLaws(t *testing.T) {
	t.Parallel()
	p := Predicate[int](func(value int) bool { return value%2 == 0 })
	values := []int{-2, 0, 2}
	if None(values, p) != !Any(values, p) {
		t.Fatal("None(xs,p) != !Any(xs,p)")
	}
	if All(values, p) != !Any(values, Not(p)) {
		t.Fatal("All(xs,p) != !Any(xs,Not(p))")
	}
}

var benchmarkQuantifierResult bool

func BenchmarkAll(b *testing.B) {
	b.ReportAllocs()
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	p := Predicate[int](func(v int) bool { return v > 0 })
	for b.Loop() {
		benchmarkQuantifierResult = All(values, p)
	}
}

func BenchmarkAnyShortCircuit(b *testing.B) {
	b.ReportAllocs()
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	p := Predicate[int](func(v int) bool { return v == 1 })
	for b.Loop() {
		benchmarkQuantifierResult = Any(values, p)
	}
}
