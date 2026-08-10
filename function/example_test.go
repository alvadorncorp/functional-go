package function_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/function"
)

func ExampleCompose() {
	length := func(value string) int { return len(value) }
	double := func(value int) int { return value * 2 }
	fmt.Println(function.Compose(double, length)("abcd"))
	// Output: 8
}
