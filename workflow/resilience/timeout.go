package resilience

import (
	"context"
	"time"

	"go.alvadorn.org/functional/workflow/action"
)

// Timeout returns a lazy Action that gives source a cooperative child deadline.
// It does not start a goroutine or preempt a source that ignores its context.
func Timeout[T any](duration time.Duration, source action.Action[T]) action.Action[T] {
	if duration <= 0 {
		panic(panicInvalidTimeout)
	}
	if source == nil {
		panic(panicNilAction)
	}

	return func(ctx context.Context) (T, error) {
		child, cancel := context.WithTimeout(ctx, duration)
		defer cancel()
		return source(child)
	}
}
