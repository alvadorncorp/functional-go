package action

import "context"

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
