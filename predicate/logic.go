package predicate

// Not returns the Boolean negation of predicate.
func Not[T any](predicate Predicate[T]) Predicate[T] {
	return func(value T) bool { return !predicate(value) }
}

// And returns a predicate that succeeds when every predicate succeeds.
// With no predicates, And returns the always-true identity predicate.
func And[T any](predicates ...Predicate[T]) Predicate[T] {
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
func Or[T any](predicates ...Predicate[T]) Predicate[T] {
	return func(value T) bool {
		for _, predicate := range predicates {
			if predicate(value) {
				return true
			}
		}
		return false
	}
}
