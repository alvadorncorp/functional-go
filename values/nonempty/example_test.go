package nonempty_test

import (
	"fmt"

	"go.alvadorn.org/functional/values/nonempty"
)

func ExampleOf() {
	values := nonempty.Of("first", "second", "third")
	fmt.Println(values.Head(), values.Last(), values.Len())
	// Output:
	// first third 3
}
