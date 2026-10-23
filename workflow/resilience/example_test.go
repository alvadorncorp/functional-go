package resilience_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alvadorncorp/gopackages/workflow/action"
	"github.com/alvadorncorp/gopackages/workflow/resilience"
)

func ExampleRetry() {
	unavailable := errors.New("unavailable")
	attempts := 0
	source := action.Do(func(context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", unavailable
		}
		return "created", nil
	})
	policy := resilience.RetryPolicy{
		Attempts: 3,
		When:     func(err error) bool { return errors.Is(err, unavailable) },
		Backoff:  resilience.Backoff(func(int) time.Duration { return 0 }),
	}

	value, err := action.Run(context.Background(), resilience.Retry(policy, source))
	fmt.Println(value, err, attempts)
	// Output: created <nil> 3
}

// ExampleTimeout_totalBudget shows Timeout around Retry: one shared budget for
// all attempts and backoffs.
func ExampleTimeout_totalBudget() {
	source := action.Pure("ok")
	policy := resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { return 0 },
	}
	program := resilience.Timeout(time.Second, resilience.Retry(policy, source))
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output: ok <nil>
}

// ExampleTimeout_perAttemptBudget shows Retry around Timeout: each attempt
// receives a fresh deadline.
func ExampleTimeout_perAttemptBudget() {
	source := action.Pure("ok")
	policy := resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { return 0 },
	}
	program := resilience.Retry(policy, resilience.Timeout(time.Second, source))
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err)
	// Output: ok <nil>
}

// ExampleRetry_perAttemptTimeout retries an expired child deadline while the
// parent context still permits another attempt.
func ExampleRetry_perAttemptTimeout() {
	attempts := 0
	source := action.Do(func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			<-ctx.Done()
			return "", ctx.Err()
		}
		return "created", nil
	})
	policy := resilience.RetryPolicy{
		Attempts: 3,
		When:     func(err error) bool { return errors.Is(err, context.DeadlineExceeded) },
	}
	program := resilience.Retry(policy, resilience.Timeout(time.Millisecond, source))
	value, err := action.Run(context.Background(), program)
	fmt.Println(value, err, attempts)
	// Output: created <nil> 3
}
