package predicate

// EqualTo returns a predicate that compares values with ==.
func EqualTo[T comparable](expected T) Predicate[T] {
	return func(value T) bool { return value == expected }
}

// NotEqualTo returns a predicate that compares values with !=.
func NotEqualTo[T comparable](expected T) Predicate[T] {
	return func(value T) bool { return value != expected }
}

// OneOf returns a predicate that performs a linear membership scan.
func OneOf[T comparable](values ...T) Predicate[T] {
	return func(value T) bool {
		for _, candidate := range values {
			if value == candidate {
				return true
			}
		}
		return false
	}
}
