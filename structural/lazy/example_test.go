package lazy_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/structural/lazy"
)

func ExampleMap() {
	config := lazy.New(func() string { return "production" })
	length := lazy.Map(config, func(s string) int { return len(s) })
	fmt.Println(length.Get())
	// Output: 10
}
