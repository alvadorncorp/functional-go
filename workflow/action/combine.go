package action

import "context"

// Pair is the minimal structural result returned by Zip. It intentionally has
// no methods or value-algebra behavior.
type Pair[A, B any] struct {
	First  A
	Second B
}

// Zip runs first and then second, preserving both results. It is sequential:
// second does not run when first fails. Zip panics with
// "workflow/action: nil action" when either argument is nil.
func Zip[A, B any](first Action[A], second Action[B]) Action[Pair[A, B]] {
	if first == nil || second == nil {
		panic(panicNilAction)
	}
	return func(ctx context.Context) (Pair[A, B], error) {
		firstValue, err := first(ctx)
		if err != nil {
			return Pair[A, B]{}, err
		}
		secondValue, err := second(ctx)
		if err != nil {
			return Pair[A, B]{}, err
		}
		return Pair[A, B]{First: firstValue, Second: secondValue}, nil
	}
}

// Both runs first and then second before combining their values. It is
// sequential and combine runs exactly once only after both Actions succeed.
// Both panics with "workflow/action: nil action" for a nil Action and
// "workflow/action: nil combiner" for a nil combine function.
func Both[A, B, C any](first Action[A], second Action[B], combine func(A, B) C) Action[C] {
	if first == nil || second == nil {
		panic(panicNilAction)
	}
	if combine == nil {
		panic(panicNilCombiner)
	}
	return func(ctx context.Context) (C, error) {
		firstValue, err := first(ctx)
		if err != nil {
			var zero C
			return zero, err
		}
		secondValue, err := second(ctx)
		if err != nil {
			var zero C
			return zero, err
		}
		return combine(firstValue, secondValue), nil
	}
}

// Traverse binds each item to an Action and runs those Actions one after
// another in index order, collecting their values in the same order.
//
// Effects interleave per item: binder is called for an item only after the
// previous item's Action has succeeded, and that item's Action runs before the
// next item is bound. Every Action observes the context Run received. Nothing
// runs until Run, and each Run traverses the items again from the start.
//
// Traverse copies items at construction, so later writes to the caller's slice
// do not change what runs; the copy is shallow, so pointer elements still share
// what they point to. Success returns a fresh, non-nil slice of len(items), which
// is empty when items is nil or empty; it is never shared between runs. The
// first failure stops the traversal: binder is not called for the remaining
// items, and Traverse returns a nil slice with the error unchanged. Traverse does
// not check the context between items; cancellation is observed only by the
// bound Actions.
//
// Traverse panics with "workflow/action: nil binder" at construction when
// binder is nil, and with "workflow/action: nil action" at execution when
// binder returns nil.
func Traverse[A, B any](items []A, binder func(A) Action[B]) Action[[]B] {
	if binder == nil {
		panic(panicNilBinder)
	}
	snapshot := append([]A(nil), items...)
	return func(ctx context.Context) ([]B, error) {
		values := make([]B, 0, len(snapshot))
		for _, item := range snapshot {
			next := binder(item)
			if next == nil {
				panic(panicNilAction)
			}
			value, err := next(ctx)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	}
}
