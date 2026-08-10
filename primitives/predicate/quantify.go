package predicate

import "slices"

// All reports whether every value satisfies predicate. Empty values return true.
func All[T any](values []T, predicate Predicate[T]) bool {
	for _, value := range values {
		if !predicate(value) {
			return false
		}
	}
	return true
}

// Any reports whether at least one value satisfies predicate. Empty values return false.
func Any[T any](values []T, predicate Predicate[T]) bool {
	return slices.ContainsFunc(values, predicate)
}

// None reports whether no value satisfies predicate. Empty values return true.
func None[T any](values []T, predicate Predicate[T]) bool {
	return !Any(values, predicate)
}
