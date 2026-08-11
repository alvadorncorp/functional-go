package action_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

func Example() {
	program := action.Map(action.Pure(2), func(value int) int { return value * 3 })

	result, err := action.Run(context.Background(), program)
	fmt.Println(result, err)
	// Output: 6 <nil>
}

func ExampleAction_deferredExecution() {
	effects := 0
	program := action.Do(func(context.Context) (string, error) {
		effects++
		return "loaded", nil
	})

	fmt.Println("effects after construction:", effects)

	value, _ := action.Run(context.Background(), program)
	fmt.Println("value:", value, "effects after Run:", effects)
	// Output:
	// effects after construction: 0
	// value: loaded effects after Run: 1
}

func ExampleRecover() {
	notFound := errors.New("account not found")

	program := action.Recover(action.Fail[string](notFound), func(err error) action.Action[string] {
		return action.Pure("fallback: " + err.Error())
	})

	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output: fallback: account not found <nil>
}
