package resilience

import (
	"context"
	"time"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

// Retry returns a lazy Action that repeats source according to policy.
// Panics from source, When, and Backoff propagate unchanged.
func Retry[T any](policy RetryPolicy, source action.Action[T]) action.Action[T] {
	validateRetry(policy, source)
	return func(ctx context.Context) (T, error) {
		return runRetry(ctx, policy, source)
	}
}

func validateRetry[T any](policy RetryPolicy, source action.Action[T]) {
	if policy.Attempts <= 0 {
		panic(panicInvalidAttempts)
	}
	if policy.When == nil {
		panic(panicNilPredicate)
	}
	if source == nil {
		panic(panicNilAction)
	}
}

func runRetry[T any](ctx context.Context, policy RetryPolicy, source action.Action[T]) (T, error) {
	var zero T

	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		value, err := source(ctx)
		if err == nil {
			return value, nil
		}
		if attempt == policy.Attempts || !policy.When(err) {
			return zero, err
		}
		if err := pauseAfterFailure(ctx, policy.Backoff, attempt); err != nil {
			return zero, err
		}
	}
}

func pauseAfterFailure(ctx context.Context, backoff Backoff, attempt int) error {
	if backoff == nil {
		return nil
	}
	delay := backoff(attempt)
	if delay < 0 {
		panic(panicNegativeBackoff)
	}
	return wait(ctx, delay)
}

func wait(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
