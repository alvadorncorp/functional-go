package action

import "context"

// Map applies mapper to the value action produces, leaving failures untouched.
//
// Neither action nor mapper runs until Run; both then observe the context Run
// received. When action fails, mapper is not called and the error is returned
// unchanged. Map panics with "workflow/action: nil action" or
// "workflow/action: nil mapper" when its arguments are nil.
func Map[A, B any](action Action[A], mapper func(A) B) Action[B] {
	if action == nil {
		panic(panicNilAction)
	}
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(ctx context.Context) (B, error) {
		value, err := action(ctx)
		if err != nil {
			var zero B
			return zero, err
		}
		return mapper(value), nil
	}
}

// MapFallible applies a failing pure transformation to the value action
// produces. It is the bridge for a pure domain function shaped func(A) (B,
// error), which stays an ordinary function instead of becoming an Action.
//
// Neither action nor mapper runs until Run. When action fails, mapper is not
// called; when mapper fails, its error is returned unchanged alongside the zero
// value of B. MapFallible panics with "workflow/action: nil action" or
// "workflow/action: nil mapper" when its arguments are nil.
func MapFallible[A, B any](action Action[A], mapper func(A) (B, error)) Action[B] {
	if action == nil {
		panic(panicNilAction)
	}
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(ctx context.Context) (B, error) {
		value, err := action(ctx)
		if err != nil {
			var zero B
			return zero, err
		}
		return mapper(value)
	}
}

// FlatMap sequences a dependent Action after action, running binder only on
// success and running the Action binder returns with the same context.
//
// Execution order is deterministic: action first, then the bound Action. When
// action fails, binder is not called and the error is returned unchanged.
// FlatMap panics with "workflow/action: nil action" or
// "workflow/action: nil binder" at construction, and with
// "workflow/action: nil action" at execution when binder returns nil.
func FlatMap[A, B any](action Action[A], binder func(A) Action[B]) Action[B] {
	if action == nil {
		panic(panicNilAction)
	}
	if binder == nil {
		panic(panicNilBinder)
	}
	return func(ctx context.Context) (B, error) {
		value, err := action(ctx)
		if err != nil {
			var zero B
			return zero, err
		}
		next := binder(value)
		if next == nil {
			panic(panicNilAction)
		}
		return next(ctx)
	}
}
