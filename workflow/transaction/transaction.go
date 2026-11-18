// Package transaction models a transactional boundary supplied by an adapter
// and the scoped capabilities that boundary provides to a workflow.
//
// The package contains no transaction implementation, names no concrete scope
// type, and holds no state. It never stores a transaction handle in a
// context.Context and offers no way to look one up: there is no
// CurrentTransaction, no TransactionFromContext, and no default boundary. A
// boundary is an ordinary capability, injected at construction like any other.
//
// A Boundary makes no atomicity claim of its own — the adapter that implements
// it documents its guarantee. A PostgreSQL boundary provides one local ACID
// transaction; an in-memory fake provides a test simulation. Composing two
// boundaries over two independent resource managers is composition of effects,
// never a distributed transaction, and this package deliberately offers no
// combinator that would suggest otherwise.
//
// Nested use has no special meaning here. Calling Within inside a boundary does
// not create a savepoint, a nested transaction, or a distributed one; whatever
// happens is entirely the adapter's semantics.
package transaction

import (
	"context"

	"go.alvadorn.org/functional/workflow/action"
)

// Panic messages are stable, unexported constants.
const (
	panicNilBoundary = "workflow/transaction: nil boundary"
	panicNilFunction = "workflow/transaction: nil function"
	panicSkippedFn   = "workflow/transaction: boundary returned without running fn"
)

// Boundary is a region of execution with transactional semantics defined by its
// provider. It runs the supplied function with a scope S of capabilities bound
// to one concrete transaction, and decides — from the error the function
// returns and from any panic that escapes it — whether that transaction is
// committed or rolled back.
//
// Implementations begin the transaction, build S from it, commit when the
// function returns nil, and roll back on a non-nil error or a panic. The
// concrete handle stays inside the implementation: S carries capabilities, not
// a driver type, which is what keeps a workflow from accidentally escaping the
// transaction.
//
// A successful boundary must invoke the supplied function exactly once, and
// that call must complete synchronously, before the boundary returns nil. It
// may return an error without invoking the function, for example when beginning
// the transaction fails.
//
// The zero value is a nil function; Within rejects it.
type Boundary[S any] func(context.Context, func(context.Context, S) error) error

// Within runs fn inside boundary and carries fn's typed result out through the
// boundary's error-only callback.
//
// Nothing runs until the returned Action is run: constructing Within begins no
// transaction. On each run the boundary produces a fresh scope, fn builds an
// Action from it, and that Action executes with the context the boundary passes
// down — so cancellation and deadlines reach the adapter normally. On success
// the value is returned and the boundary commits; on failure the zero value and
// the error are returned unchanged and the boundary rolls back. A panic inside
// fn is not converted into an error: it unwinds through the boundary, which is
// responsible for attempting a rollback before it propagates.
//
// The boundary must invoke its callback exactly once, completing it before
// returning nil; returning an error without invoking the callback returns that
// error unchanged. Within neither counts nor prevents repeated callback calls:
// those remain a violation of the adapter contract.
//
// Within is a free function rather than a method because Go does not allow a
// method to introduce the extra type parameter T, and preserving T beats a
// fluent shape that would need any or reflection.
//
// Within panics with "workflow/transaction: nil boundary" or
// "workflow/transaction: nil function" at construction, and with
// "workflow/action: nil action" at execution when fn returns a nil Action.
// It also panics with "workflow/transaction: boundary returned without running fn"
// when the boundary returns nil without invoking its callback.
func Within[S, T any](boundary Boundary[S], fn func(S) action.Action[T]) action.Action[T] {
	if boundary == nil {
		panic(panicNilBoundary)
	}
	if fn == nil {
		panic(panicNilFunction)
	}
	return action.Do(func(ctx context.Context) (T, error) {
		var result T
		var ran bool

		err := boundary(ctx, func(ctx context.Context, scope S) error {
			ran = true
			value, err := action.Run(ctx, fn(scope))
			if err != nil {
				return err
			}
			result = value
			return nil
		})
		if err != nil {
			var zero T
			return zero, err
		}
		if !ran {
			panic(panicSkippedFn)
		}
		return result, nil
	})
}
