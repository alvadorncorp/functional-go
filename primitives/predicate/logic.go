package predicate

import "slices"

// Not returns the Boolean negation of predicate.
func Not[T any](predicate Predicate[T]) Predicate[T] {
	return func(value T) bool { return !predicate(value) }
}

// And returns a predicate that succeeds when every predicate succeeds.
// With no predicates, And returns the always-true identity predicate.
// And copies the variadic slice; later changes to it do not affect the result.
func And[T any](predicates ...Predicate[T]) Predicate[T] {
	predicates = slices.Clone(predicates)
	return func(value T) bool {
		for _, predicate := range predicates {
			if !predicate(value) {
				return false
			}
		}
		return true
	}
}

// Or returns a predicate that succeeds when any predicate succeeds.
// With no predicates, Or returns the always-false identity predicate.
// Or copies the variadic slice; later changes to it do not affect the result.
func Or[T any](predicates ...Predicate[T]) Predicate[T] {
	predicates = slices.Clone(predicates)
	return func(value T) bool {
		for _, predicate := range predicates {
			if predicate(value) {
				return true
			}
		}
		return false
	}
}
