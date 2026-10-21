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

func ExampleFromPair() {
	parsed := result.FromPair(strconv.Atoi("42"))
	fmt.Println(parsed.Get())
	fmt.Println(result.FromPair(strconv.Atoi("forty-two")).IsErr())
	value, err := result.ToPair(parsed)
	fmt.Println(value, err)
	// Output:
	// 42 true
	// true
	// 42 <nil>
}

func ExampleMatch() {
	describe := func(r result.Result[int, error]) string {
		return result.Match(
			r,
			func(err error) string { return "failed: " + err.Error() },
			func(v int) string { return "parsed " + strconv.Itoa(v) },
		)
	}
	fmt.Println(describe(result.FromPair(strconv.Atoi("42"))))
	fmt.Println(describe(result.Err[int](strconv.ErrSyntax)))
	// Output:
	// parsed 42
	// failed: invalid syntax
}
