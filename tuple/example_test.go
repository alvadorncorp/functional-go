package tuple_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/tuple"
)

func ExampleOf2() {
	pair := tuple.Of2("Ada", 36)
	fmt.Println(pair.First, pair.Second)
	// Output: Ada 36
}
