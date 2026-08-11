package action

import "context"

// Bracket acquires a resource, uses it, and releases it after every normal use
// return. Release runs if and only if acquire succeeds. The same context reaches
// all three stages; panics are not recovered and do not promise release.
func Bracket[R, T any](
	acquire Action[R],
	use func(R) Action[T],
	release func(R) Action[struct{}],
) Action[T] {
	if acquire == nil {
		panic(panicNilAction)
	}
	if use == nil {
		panic(panicNilUse)
	}
	if release == nil {
		panic(panicNilRelease)
	}
	return func(ctx context.Context) (T, error) {
		resource, err := acquire(ctx)
		if err != nil {
			var zero T
			return zero, err
		}
		useAction := use(resource)
		if useAction == nil {
			panic(panicNilAction)
		}
		value, useErr := useAction(ctx)
		releaseAction := release(resource)
		if releaseAction == nil {
			panic(panicNilAction)
		}
		_, releaseErr := releaseAction(ctx)
		return finalizeResult(value, useErr, releaseErr)
	}
}

// AcquireUseRelease is the descriptive name for Bracket and delegates to its
// single lifecycle implementation.
func AcquireUseRelease[R, T any](
	acquire Action[R],
	use func(R) Action[T],
	release func(R) Action[struct{}],
) Action[T] {
	return Bracket(acquire, use, release)
}
