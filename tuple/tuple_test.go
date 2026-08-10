package tuple_test

import (
	"reflect"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/tuple"
)

func TestTuple2ConstructionAndTransforms(t *testing.T) {
	t.Parallel()

	got := tuple.Of2("age", 42)
	if got.First != "age" || got.Second != 42 {
		t.Fatalf("Of2 = %#v", got)
	}
	if from := tuple.FromValues2("age", 42); from != got {
		t.Fatalf("FromValues2 = %#v, want %#v", from, got)
	}
	if swapped := tuple.Swap(got); swapped.First != 42 || swapped.Second != "age" {
		t.Fatalf("Swap = %#v", swapped)
	}
	if mapped := tuple.MapFirst(got, func(v string) int { return len(v) }); mapped.First != 3 || mapped.Second != 42 {
		t.Fatalf("MapFirst = %#v", mapped)
	}
	mapToTens := func(v int) string { return string(rune('0' + v/10)) }
	if mapped := tuple.MapSecond(got, mapToTens); mapped.First != "age" || mapped.Second != "4" {
		t.Fatalf("MapSecond = %#v", mapped)
	}
}

func TestTuple3ConstructionAndTransforms(t *testing.T) {
	t.Parallel()

	got := tuple.Of3("x", 2, true)
	if got.First != "x" || got.Second != 2 || !got.Third {
		t.Fatalf("Of3 = %#v", got)
	}
	first := tuple.MapFirst3(got, func(v string) int { return len(v) })
	second := tuple.MapSecond3(got, func(v int) int64 { return int64(v * 2) })
	third := tuple.MapThird3(got, func(v bool) string {
		if v {
			return "yes"
		}
		return "no"
	})
	if first.First != 1 || first.Second != 2 || !first.Third ||
		second.First != "x" || second.Second != 4 || !second.Third ||
		third.First != "x" || third.Second != 2 || third.Third != "yes" {
		t.Fatalf("maps = %#v %#v %#v", first, second, third)
	}
}

func TestTupleZeroPointersAndNonComparableValues(t *testing.T) {
	t.Parallel()

	var zero2 tuple.Tuple2[int, string]
	var zero3 tuple.Tuple3[int, string, *int]
	if zero2.First != 0 || zero2.Second != "" || zero3.Third != nil {
		t.Fatalf("zero values = %#v %#v", zero2, zero3)
	}
	value := 7
	withPointer := tuple.Of2(&value, "value")
	if withPointer.First != &value || *withPointer.First != 7 {
		t.Fatalf("pointer tuple = %#v", withPointer)
	}
	withSlices := tuple.Of2([]int{1, 2}, map[string]int{"x": 1})
	if !reflect.DeepEqual(withSlices.First, []int{1, 2}) || withSlices.Second["x"] != 1 {
		t.Fatalf("non-comparable tuple = %#v", withSlices)
	}
}
