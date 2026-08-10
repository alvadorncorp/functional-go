package predicate_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/predicate"
)

func ExampleAnd() {
	positiveEven := predicate.And[int](func(v int) bool { return v > 0 }, func(v int) bool { return v%2 == 0 })
	fmt.Println(positiveEven(4), positiveEven(3))
	// Output: true false
}
