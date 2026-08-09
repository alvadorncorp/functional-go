package nonempty_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/values/nonempty"
)

func ExampleOf() {
	values := nonempty.Of("first", "second", "third")
	fmt.Println(values.Head(), values.Last(), values.Len())
	// Output:
	// first third 3
}
