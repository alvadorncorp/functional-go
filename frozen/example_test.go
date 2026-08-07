package frozen_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/frozen"
)

// ExampleNewSlice shows that NewSlice copies its input: mutating the source
// slice after construction does not reach the Slice.
func ExampleNewSlice() {
	xs := []int{1, 2, 3}
	s := frozen.NewSlice(xs...)

	xs[0] = -1 // does not reach s

	fmt.Println(s.Items())
	// Output:
	// [1 2 3]
}

// ExampleSlice_Values shows the allocation-free way to read a Slice.
func ExampleSlice_Values() {
	s := frozen.NewSlice(1, 2, 3)

	sum := 0
	for v := range s.Values() {
		sum += v
	}

	fmt.Println(sum)
	// Output:
	// 6
}
