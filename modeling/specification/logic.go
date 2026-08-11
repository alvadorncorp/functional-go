package specification

// All returns a Specification that accepts when every specification accepts.
// With no operands, All returns the always-true identity.
func All[T any](specifications ...Specification[T]) Specification[T] {
	requireSpecifications(specifications)
	return func(value T) bool {
		for _, spec := range specifications {
			if !spec(value) {
				return false
			}
		}
		return true
	}
}

// Any returns a Specification that accepts when any specification accepts.
// With no operands, Any returns the always-false identity.
func Any[T any](specifications ...Specification[T]) Specification[T] {
	requireSpecifications(specifications)
	return func(value T) bool {
		for _, spec := range specifications {
			if spec(value) {
				return true
			}
		}
		return false
	}
}

// Not returns the Boolean negation of spec.
func Not[T any](spec Specification[T]) Specification[T] {
	requireSpecifications([]Specification[T]{spec})
	return func(value T) bool { return !spec(value) }
}
