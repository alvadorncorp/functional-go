package iterator_test

import (
	"fmt"
	"slices"

	"github.com/alvadorncorp/gopackages/functional/structural/iterator"
)

func ExampleMap() {
	seq := iterator.Map(
		iterator.Filter(slices.Values([]int{1, 2, 3, 4}), func(v int) bool { return v%2 == 0 }),
		func(v int) int { return v * 10 },
	)
	fmt.Println(slices.Collect(seq))
	// Output: [20 40]
}
