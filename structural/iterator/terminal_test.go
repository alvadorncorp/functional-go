package iterator_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/structural/iterator"
	"github.com/alvadorncorp/gopackages/functional/structural/tuple"
)

func tracked(values []int, yielded, stopped *int) iter.Seq[int] {
	return func(yield func(int) bool) {
		defer func() { (*stopped)++ }()
		for _, value := range values {
			(*yielded)++
			if !yield(value) {
				return
			}
		}
	}
}

func TestZipAndEnumerate(t *testing.T) {
	t.Parallel()
	leftYielded, rightYielded, leftStopped, rightStopped := 0, 0, 0, 0
	left := tracked([]int{1, 2, 3}, &leftYielded, &leftStopped)
	right := tracked([]int{10, 20}, &rightYielded, &rightStopped)
	got := slices.Collect(iterator.Zip(left, right))
	want := []tuple.Tuple2[int, int]{tuple.Of2(1, 10), tuple.Of2(2, 20)}
	if !slices.Equal(got, want) {
		t.Fatalf("Zip = %v", got)
	}
	if leftYielded != 3 || rightYielded != 2 || leftStopped != 1 || rightStopped != 1 {
		t.Fatalf("yielded=%d/%d stopped=%d/%d", leftYielded, rightYielded, leftStopped, rightStopped)
	}
	enumerated := slices.Collect(iterator.Enumerate(slices.Values([]string{"a", "b"})))
	wantEnumerated := []tuple.Tuple2[int, string]{tuple.Of2(0, "a"), tuple.Of2(1, "b")}
	if !slices.Equal(enumerated, wantEnumerated) {
		t.Fatalf("Enumerate = %v", enumerated)
	}
}

func TestZipStopsBothSourcesOnDownstreamTermination(t *testing.T) {
	t.Parallel()
	leftYielded, rightYielded, leftStopped, rightStopped := 0, 0, 0, 0
	left := tracked([]int{1, 2, 3}, &leftYielded, &leftStopped)
	right := tracked([]int{10, 20, 30}, &rightYielded, &rightStopped)
	got := slices.Collect(iterator.Take(iterator.Zip(left, right), 1))
	if len(got) != 1 || got[0] != tuple.Of2(1, 10) {
		t.Fatalf("Zip Take(1) = %v", got)
	}
	if leftYielded != 1 || rightYielded != 1 || leftStopped != 1 || rightStopped != 1 {
		t.Fatalf(
			"yielded=%d/%d stopped=%d/%d",
			leftYielded, rightYielded, leftStopped, rightStopped,
		)
	}
}

func TestTerminalsAndShortCircuit(t *testing.T) {
	t.Parallel()
	seq := slices.Values([]int{1, 2, 3, 4})
	if got := iterator.Fold(seq, 0, func(total, value int) int { return total + value }); got != 10 {
		t.Fatalf("Fold = %d", got)
	}
	yielded := 0
	if !iterator.Any(counted([]int{1, 2, 3}, &yielded), func(v int) bool { return v == 2 }) || yielded != 2 {
		t.Fatalf("Any yielded=%d", yielded)
	}
	if iterator.Any(iterator.Empty[int](), func(int) bool { return true }) {
		t.Fatal("Any(empty)")
	}
	if !iterator.All(iterator.Empty[int](), func(int) bool { return false }) {
		t.Fatal("All(empty)")
	}
	yielded = 0
	if iterator.All(counted([]int{1, 2, 3}, &yielded), func(v int) bool { return v < 2 }) || yielded != 2 {
		t.Fatalf("All yielded=%d", yielded)
	}
	yielded = 0
	match, found := iterator.Find(
		counted([]int{1, 2, 3, 4}, &yielded),
		func(v int) bool { return v > 2 },
	)
	if !found || match != 3 || yielded != 3 {
		t.Fatalf("Find = %d,%v yielded=%d", match, found, yielded)
	}
	yielded = 0
	if got, ok := iterator.First(counted([]int{1, 2, 3}, &yielded)); !ok || got != 1 || yielded != 1 {
		t.Fatalf("First = %d,%v yielded=%d", got, ok, yielded)
	}
	if got, ok := iterator.Last(seq); !ok || got != 4 {
		t.Fatalf("Last = %d,%v", got, ok)
	}
	if _, ok := iterator.Last(iterator.Empty[int]()); ok {
		t.Fatal("Last(empty)")
	}
	if got := iterator.Count(seq); got != 4 {
		t.Fatalf("Count = %d", got)
	}
}

func TestFindNotFound(t *testing.T) {
	t.Parallel()
	yielded, checked := 0, 0
	got, found := iterator.Find(counted([]int{1, 2, 3}, &yielded), func(v int) bool {
		checked++
		return v > 3
	})
	if got != 0 || found || yielded != 3 || checked != 3 {
		t.Fatalf("Find = %d,%v yielded=%d checked=%d", got, found, yielded, checked)
	}
}

func TestFirstEmpty(t *testing.T) {
	t.Parallel()
	yielded := 0
	got, found := iterator.First(counted(nil, &yielded))
	if got != 0 || found || yielded != 0 {
		t.Fatalf("First(empty) = %d,%v yielded=%d", got, found, yielded)
	}
}

func TestInfiniteSourcesAndUnfold(t *testing.T) {
	t.Parallel()
	if got := slices.Collect(iterator.Take(iterator.Repeat("x"), 3)); !slices.Equal(got, []string{"x", "x", "x"}) {
		t.Fatalf("Repeat = %v", got)
	}
	iterated := iterator.Take(iterator.Iterate(1, func(v int) int { return v * 2 }), 4)
	if got := slices.Collect(iterated); !slices.Equal(got, []int{1, 2, 4, 8}) {
		t.Fatalf("Iterate = %v", got)
	}
	seq := iterator.Unfold(0, func(state int) (int, int, bool) { return state, state + 1, state < 3 })
	if got := slices.Collect(seq); !slices.Equal(got, []int{0, 1, 2}) {
		t.Fatalf("Unfold = %v", got)
	}
}
