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
