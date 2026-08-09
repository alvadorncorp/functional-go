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
