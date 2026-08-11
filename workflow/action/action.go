// Package action provides Action, a deferred and context-bound computation
// that may fail, together with the combinators that compose Actions without
// executing them.
//
// Building an Action performs no work and starts no goroutine. Effects happen
// only when Run is called, and the context Run receives is the context every
// derived Action observes. An Action is re-executable rather than memoized:
// running the same Action twice runs its effect twice. It is neither a future
// nor a promise, and nothing in this package begins work in the background.
//
// Single errors travel unchanged, so errors.Is and errors.As keep working and
// context.Canceled and context.DeadlineExceeded reach the caller unchanged.
// Ensure and Bracket deliberately use errors.Join only when independent primary
// and finalizer errors both exist. Panics are never converted into errors.
package action

import "context"

// Action is a computation that produces a T when executed with a context.
//
// The zero value is a nil function and is not runnable; Run panics with
// "workflow/action: nil action" rather than deferring to the runtime's nil
// function panic. An Action value is safe for concurrent use and safe to reuse
// whenever the dependencies its closure captured are, because the package adds
// no mutable state of its own.
type Action[T any] func(context.Context) (T, error)

// Run executes action with ctx and is the single passage from description to
// execution in this package.
//
// ctx is passed to action unchanged and, through it, to every derived Action.
// The returned error is whatever action produced, never wrapped. Run panics
// with "workflow/action: nil action" when action is nil.
func Run[T any](ctx context.Context, action Action[T]) (T, error) {
	if action == nil {
		panic(panicNilAction)
	}
	return action(ctx)
}
