package predicate

import "slices"

// All reports whether every value satisfies predicate. Empty values return true.
//
// Deprecated: Use !slices.ContainsFunc(values, func(value T) bool { return !predicate(value) }).
func All[T any](values []T, predicate Predicate[T]) bool {
	for _, value := range values {
		if !predicate(value) {
			return false
		}
	}
	return true
}

// Any reports whether at least one value satisfies predicate. Empty values return false.
//
// Deprecated: Use slices.ContainsFunc(values, predicate).
func Any[T any](values []T, predicate Predicate[T]) bool {
	return slices.ContainsFunc(values, predicate)
}

// None reports whether no value satisfies predicate. Empty values return true.
//
// Deprecated: Use !slices.ContainsFunc(values, predicate).
func None[T any](values []T, predicate Predicate[T]) bool {
	return !Any(values, predicate)
}
