package nonempty_test

import (
	"testing"

	"go.alvadorn.org/functional/values/nonempty"
)

func TestNonEmptyZeroAndConstructors(t *testing.T) {
	t.Parallel()
	var zero nonempty.NonEmpty[int]
	if zero.Len() != 1 || zero.Head() != 0 || zero.Last() != 0 || zero.At(0) != 0 {
		t.Fatalf("zero = len %d head %d last %d", zero.Len(), zero.Head(), zero.Last())
	}
	if got := zero.Items(); len(got) != 1 || got[0] != 0 {
		t.Fatalf("zero.Items() = %v", got)
	}

	one := nonempty.Of(1)
	if one.Len() != 1 || one.Head() != 1 || one.Last() != 1 {
		t.Fatalf("Of(1) = %v", one.Items())
	}
	many := nonempty.Of(1, 2, 3)
	if many.Len() != 3 || many.At(1) != 2 || many.Last() != 3 {
		t.Fatalf("Of(1,2,3) = %v", many.Items())
	}
	if _, ok := nonempty.FromSlice[int](nil); ok {
		t.Fatal("FromSlice(nil) must fail")
	}
	if _, ok := nonempty.FromSlice([]int{}); ok {
		t.Fatal("FromSlice(empty) must fail")
	}
	from, ok := nonempty.FromSlice([]int{4, 5})
	if !ok || from.Head() != 4 || from.Last() != 5 {
		t.Fatalf("FromSlice = %v, %v", from.Items(), ok)
	}
}

func TestFromSliceDefensiveCopy(t *testing.T) {
	t.Parallel()
	src := []int{1, 2, 3}
	n, ok := nonempty.FromSlice(src)
	if !ok {
		t.Fatal("FromSlice failed")
	}
	src[0], src[1] = 99, 99
	if got := n.Items(); got[0] != 1 || got[1] != 2 {
		t.Fatalf("caller mutation changed value: %v", got)
	}

	items := n.Items()
	items[0] = 88
	tail := n.Tail()
	tail[0] = 77
	if got := n.Items(); got[0] != 1 || got[1] != 2 {
		t.Fatalf("returned-slice mutation changed value: %v", got)
	}
}

func TestOfDefensiveCopyOfCallerTail(t *testing.T) {
	t.Parallel()
	tail := []int{2, 3}
	n := nonempty.Of(1, tail...)
	tail[0], tail[1] = 99, 99

	if got := n.Items(); got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("caller mutation changed value: %v", got)
	}
}

func TestMapAndAppendPreserveValueSemantics(t *testing.T) {
	t.Parallel()
	n := nonempty.Of(1, 2)
	mapped := nonempty.Map(n, func(v int) string { return string(rune('0' + v)) })
	if got := mapped.Items(); len(got) != 2 || got[0] != "1" || got[1] != "2" {
		t.Fatalf("Map = %v", got)
	}
	appended := n.Append(3, 4)
	if got := appended.Items(); len(got) != 4 || got[3] != 4 {
		t.Fatalf("Append = %v", got)
	}
	if got := n.Items(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("Append mutated source: %v", got)
	}
}

func TestAppendDefensiveCopyOfCallerValues(t *testing.T) {
	t.Parallel()
	values := []int{3, 4}
	appended := nonempty.Of(1, 2).Append(values...)
	values[0], values[1] = 99, 99

	if got := appended.Items(); got[0] != 1 || got[1] != 2 || got[2] != 3 || got[3] != 4 {
		t.Fatalf("caller mutation changed appended value: %v", got)
	}
}

func TestAtPanicsOutsideBounds(t *testing.T) {
	t.Parallel()
	for _, index := range []int{-1, 2} {
		t.Run("bounds", func(t *testing.T) {
			defer func() {
				if got := recover(); got != "functional/values/nonempty: index out of range" {
					t.Fatalf("panic = %v", got)
				}
			}()
			nonempty.Of(1, 2).At(index)
		})
	}
}
