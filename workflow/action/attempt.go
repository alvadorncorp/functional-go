package action

import "context"

// AttemptResult is the normal value/error pair observed by Attempt. It is a
// minimal carrier and intentionally has no methods.
type AttemptResult[T any] struct {
	Value T
	Err   error
}

// Attempt turns the normal error channel of source into data. Its returned
// Action has a nil outer error after every normal source return and preserves
// both the value and error source returned. Panics propagate unchanged.
// Attempt panics with "workflow/action: nil action" when source is nil.
func Attempt[T any](source Action[T]) Action[AttemptResult[T]] {
	if source == nil {
		panic(panicNilAction)
	}
	return func(ctx context.Context) (AttemptResult[T], error) {
		value, err := source(ctx)
		return AttemptResult[T]{Value: value, Err: err}, nil
	}
}
