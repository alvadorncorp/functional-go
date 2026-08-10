package order_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/primitives/order"
)

func ExampleBy() {
	byAge := order.By(func(value person) int { return value.age }, order.Natural[int]())
	fmt.Println(order.Less(byAge)(person{age: 20}, person{age: 30}))
	// Output: true
}
