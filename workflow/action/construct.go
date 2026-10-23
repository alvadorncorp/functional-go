package action

import "context"

// Do lifts an ordinary Go function into an Action.
//
// fn is not called by Do; it runs once per Run, with the context Run received.
// This is the escape hatch that keeps the package from becoming a closed DSL:
// arbitrary imperative Go inside fn stays composable with every combinator
// here. Do panics with "workflow/action: nil function" when fn is nil.
func Do[T any](fn func(context.Context) (T, error)) Action[T] {
	if fn == nil {
		panic(panicNilFunction)
	}
	return fn
}

// Exec lifts an error-only Go function into an Action with no result value.
//
// Like Do, Exec does not call fn during construction; fn runs once per Run
// with the same context. Its error and any panic propagate unchanged.
// Exec panics with "workflow/action: nil function" when fn is nil.
func Exec(fn func(context.Context) error) Action[struct{}] {
	if fn == nil {
		panic(panicNilFunction)
	}
	return func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	}
}

// Pure is an Action that yields value and never fails.
//
// It executes no effect, ignores its context, and returns the same value on
// every Run. Nil is not rejected because value is data, not a callback.
func Pure[T any](value T) Action[T] {
	return func(context.Context) (T, error) {
		return value, nil
	}
}

// Fail is an Action that always fails with err, preserving its identity so
// errors.Is and errors.As work on the result of Run.
//
// It executes no effect and ignores its context. Fail panics with
// "workflow/action: nil error" when err is nil, because a failure Action with
// no failure is a programming defect.
func Fail[T any](err error) Action[T] {
	if err == nil {
		panic(panicNilError)
	}
	return func(context.Context) (T, error) {
		var zero T
		return zero, err
	}
}
