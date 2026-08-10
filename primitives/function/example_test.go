package function_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/primitives/function"
)

func ExampleCompose() {
	length := func(value string) int { return len(value) }
	double := func(value int) int { return value * 2 }
	fmt.Println(function.Compose(double, length)("abcd"))
	// Output: 8
}

func ExamplePipeTransforms() {
	inc := function.Transform[int](func(value int) int { return value + 10 })
	double := function.Transform[int](func(value int) int { return value * 2 })
	transform := function.PipeTransforms(inc, double)
	fmt.Println(transform(5))
	// Output: 30
}
