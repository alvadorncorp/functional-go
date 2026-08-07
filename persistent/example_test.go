package persistent_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

// ExampleList_Append shows that Append returns a new List and leaves the
// original untouched.
func ExampleList_Append() {
	original := persistent.NewList(1, 2, 3)
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
	base := persistent.NewList(1, 2, 3)
	built := base.Builder().
		Append(4).
		Set(0, 99).
		Delete(1).
		Build()

	fmt.Println(built.Items())
	// Output:
	// [99 3 4]
}

// ExampleMap_Set shows that Set returns a new Map holding both the added
// key and every key already present.
func ExampleMap_Set() {
	base := persistent.NewMap(map[string]int{"a": 1})
	updated := base.Set("b", 2)

	a, _ := updated.Get("a")
	b, _ := updated.Get("b")
	fmt.Println(a, b)
	// Output:
	// 1 2
}
