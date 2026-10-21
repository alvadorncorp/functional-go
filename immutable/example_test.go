package immutable_test

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// ExampleListOf shows constructing a List with ListOf.
func ExampleListOf() {
	list := immutable.ListOf(1, 2, 3)
	fmt.Println(list.Items())
	fmt.Println(list.IsEmpty())
	// Output:
	// [1 2 3]
	// false
}

// ExampleList_Append shows that Append returns a new List and leaves the
// original untouched.
func ExampleList_Append() {
	original := immutable.ListOf(1, 2, 3)
	extended := original.Append(4, 5)

	fmt.Println(original.Items())
	fmt.Println(extended.Items())
	// Output:
	// [1 2 3]
	// [1 2 3 4 5]
}

// ExampleList_Builder shows the Builder path for accumulating several
// writes before producing one new, immutable List.
func ExampleList_Builder() {
	base := immutable.ListOf(1, 2, 3)
	built := base.Builder().
		Append(4).
		Set(0, 99).
		Delete(1).
		Build()

	fmt.Println(built.Items())
	// Output:
	// [99 3 4]
}

// ExampleMapOf shows constructing a Map with MapOf.
func ExampleMapOf() {
	m := immutable.MapOf(map[string]int{"a": 1, "b": 2})
	fmt.Println(m.Len())
	fmt.Println(m.Has("a"))
	// Output:
	// 2
	// true
}

// ExampleMap_Set shows that Set returns a new Map holding both the added
// key and every key already present.
func ExampleMap_Set() {
	base := immutable.MapOf(map[string]int{"a": 1})
	updated := base.Set("b", 2)

	a, _ := updated.Get("a")
	b, _ := updated.Get("b")
	fmt.Println(a, b)
	// Output:
	// 1 2
}

// ExampleMap_Update shows Update deriving a value from the old one, or from
// the zero value when the key is absent, while the original Map stays
// unchanged.
func ExampleMap_Update() {
	inc := func(old int, _ bool) int { return old + 1 }
	counts := immutable.MapOf(map[string]int{"a": 1})
	updated := counts.Update("a", inc).Update("b", inc)

	a, _ := updated.Get("a")
	b, _ := updated.Get("b")
	before, _ := counts.Get("a")
	fmt.Println(a, b, before)
	// Output:
	// 2 1 1
}

// ExampleMap_Builder shows bulk construction under exclusive ownership,
// then O(1) Build.
func ExampleMap_Builder() {
	built := immutable.MapOf[string, int](nil).Builder().
		Set("a", 1).
		Set("b", 2).
		Build()

	fmt.Println(built.Len())
	a, _ := built.Get("a")
	fmt.Println(a)
	// Output:
	// 2
	// 1
}

// Example_sortedMap shows a SortedMap ordered by cmp.Compare: iteration is
// ascending by key regardless of insertion order, and every write derives a
// new SortedMap.
func Example_sortedMap() {
	scores := immutable.EmptySortedMap[string, int](cmp.Compare[string]).
		Set("carol", 7).
		Set("alice", 9).
		Set("bob", 4)
	updated := scores.Set("alice", 10).Delete("bob")

	for name, score := range scores.All() {
		fmt.Println(name, score)
	}
	fmt.Println(slices.Collect(updated.Keys()), slices.Collect(updated.Values()))
	// Output:
	// alice 9
	// bob 4
	// carol 7
	// [alice carol] [10 7]
}

// ExampleSetOf shows constructing a Set with SetOf (deduplicated).
func ExampleSetOf() {
	s := immutable.SetOf(1, 2, 2, 3)
	fmt.Println(s.Len())
	fmt.Println(s.Has(2))
	// Output:
	// 3
	// true
}

// ExampleSet_Union shows Union, Intersection, and Difference deriving new
// Sets while both operands stay unchanged. Set iteration order is
// unspecified, so the output is sorted.
func ExampleSet_Union() {
	a := immutable.SetOf(1, 2, 3)
	b := immutable.SetOf(3, 4)

	fmt.Println(slices.Sorted(a.Union(b).Values()))
	fmt.Println(slices.Sorted(a.Intersection(b).Values()))
	fmt.Println(slices.Sorted(a.Difference(b).Values()))
	fmt.Println(slices.Sorted(a.Values()), slices.Sorted(b.Values()))
	// Output:
	// [1 2 3 4]
	// [3]
	// [1 2]
	// [1 2 3] [3 4]
}
