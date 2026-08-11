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

func ExampleZip() {
	var program action.Action[action.Pair[int, string]] = action.Zip(
		action.Pure(42),
		action.Pure("answer"),
	)
	pair, err := action.Run(context.Background(), program)
	fmt.Println(pair.First, pair.Second, err)
	// Output: 42 answer <nil>
}

func ExampleBoth() {
	program := action.Both(action.Pure(20), action.Pure(22), func(a, b int) int { return a + b })
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output: 42 <nil>
}

func ExampleTap() {
	program := action.Tap(action.Pure(42), func(value int) action.Action[struct{}] {
		fmt.Println("observed", value)
		return action.Pure(struct{}{})
	})
	value, err := action.Run(context.Background(), program)
	fmt.Println("result", value, err)
	// Output:
	// observed 42
	// result 42 <nil>
}

func ExampleAttempt() {
	wantErr := errors.New("unavailable")
	var program action.Action[action.AttemptResult[int]] = action.Attempt(action.Fail[int](wantErr))
	result, outerErr := action.Run(context.Background(), program)
	fmt.Println(result.Value, result.Err, outerErr)
	// Output: 0 unavailable <nil>
}

func ExampleEnsure() {
	program := action.Ensure(
		action.Pure(42),
		action.Do(func(context.Context) (struct{}, error) { fmt.Println("finalized"); return struct{}{}, nil }),
	)
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output:
	// finalized
	// 42 <nil>
}

func ExampleBracket() {
	program := action.Bracket(
		action.Pure("resource"),
		func(resource string) action.Action[int] { fmt.Println("use", resource); return action.Pure(42) },
		func(resource string) action.Action[struct{}] {
			fmt.Println("release", resource)
			return action.Pure(struct{}{})
		},
	)
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output:
	// use resource
	// release resource
	// 42 <nil>
}

func ExampleAcquireUseRelease() {
	program := action.AcquireUseRelease(
		action.Pure("lease"),
		func(string) action.Action[int] { return action.Pure(42) },
		func(string) action.Action[struct{}] { return action.Pure(struct{}{}) },
	)
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output: 42 <nil>
}
