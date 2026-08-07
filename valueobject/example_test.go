package valueobject_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/valueobject"
)

// ExampleCheck shows the error format Check returns: one line per
// violation, "<path>: <reason> (<type>)".
func ExampleCheck() {
	err := valueobject.Check[Bad]()
	fmt.Println(err)
	// Output:
	// Bad.Name: field is exported: a caller can assign to it directly (string)
	// Bad.tags: field can be mutated through a pointer/slice/map/chan/func alias regardless of export status ([]string)
}
