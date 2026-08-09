package either_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/values/either"
)

func ExampleMatch() {
	value := either.Right[string, int](42)
	text := either.Match(
		value,
		func(left string) string { return left },
		func(right int) string { return fmt.Sprint(right) },
	)
	fmt.Println(text)
	// Output:
	// 42
}
