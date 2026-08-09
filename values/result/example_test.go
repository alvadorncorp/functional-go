package result_test

import (
	"fmt"
	"strconv"

	"github.com/alvadorncorp/gopackages/functional/values/result"
)

func ExampleMap() {
	parsed := result.Ok[int, error](21)
	doubled := result.Map(parsed, func(v int) int { return v * 2 })
	fmt.Println(doubled.Get())
	failure := result.Err[int](strconv.ErrSyntax)
	fmt.Println(failure.IsErr())
	// Output:
	// 42 true
	// true
}
