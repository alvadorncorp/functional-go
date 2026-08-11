package action

import (
	"context"
	"errors"
)

// Tap runs effect only after source succeeds, then preserves the source value.
// The effect Action receives the same context and its error fails the result.
// Tap panics with "workflow/action: nil action" for a nil source,
// "workflow/action: nil effect" for a nil callback, and
// "workflow/action: nil action" during execution when effect returns nil.
func Tap[T any](source Action[T], effect func(T) Action[struct{}]) Action[T] {
	if source == nil {
		panic(panicNilAction)
	}
	if effect == nil {
		panic(panicNilEffect)
	}
	return func(ctx context.Context) (T, error) {
		value, err := source(ctx)
		if err != nil {
			var zero T
			return zero, err
		}
		next := effect(value)
		if next == nil {
			panic(panicNilAction)
		}
		if _, err := next(ctx); err != nil {
			var zero T
			return zero, err
		}
		return value, nil
	}
}

// Ensure runs finalizer exactly once after every normal source return. It
// forwards the same context, returns exact single errors, and joins independent
// source/finalizer errors. It does not finalize after a panic. Ensure panics
// with "workflow/action: nil action" when either argument is nil.
func Ensure[T any](source Action[T], finalizer Action[struct{}]) Action[T] {
	if source == nil || finalizer == nil {
		panic(panicNilAction)
	}
	return func(ctx context.Context) (T, error) {
		value, primaryErr := source(ctx)
		_, finalizerErr := finalizer(ctx)
		return finalizeResult(value, primaryErr, finalizerErr)
	}
}

func finalizeResult[T any](value T, primaryErr, finalizerErr error) (T, error) {
	switch {
	case primaryErr != nil && finalizerErr != nil:
		var zero T
		return zero, errors.Join(primaryErr, finalizerErr)
	case primaryErr != nil:
		var zero T
		return zero, primaryErr
	case finalizerErr != nil:
		var zero T
		return zero, finalizerErr
	default:
		return value, nil
	}
}
