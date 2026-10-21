package optional_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

func ExampleSome() {
	o := optional.Some(3)
	doubled := optional.Map(o, func(v int) int { return v * 2 })
	fmt.Println(doubled.Get())
	// Output:
	// 6 true
}

func ExampleNone() {
	var o optional.Optional[string]
	fmt.Println(o.IsNone(), o.OrElse("fallback"))
	// Output:
	// true fallback
}

func ExampleMatch() {
	greet := func(name optional.Optional[string]) string {
		return optional.Match(
			name,
			func() string { return "hello, stranger" },
			func(name string) string { return "hello, " + name },
		)
	}
	fmt.Println(greet(optional.Some("ada")))
	fmt.Println(greet(optional.None[string]()))
	// Output:
	// hello, ada
	// hello, stranger
}
