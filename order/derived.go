package order

// Less adapts order to a less-than predicate.
func Less[T any](order Order[T]) func(a, b T) bool {
	return func(a, b T) bool { return order(a, b) < 0 }
}

// Greater adapts order to a greater-than predicate.
func Greater[T any](order Order[T]) func(a, b T) bool {
	return func(a, b T) bool { return order(a, b) > 0 }
}

// Equal adapts order to an equivalence predicate.
func Equal[T any](order Order[T]) func(a, b T) bool {
	return func(a, b T) bool { return order(a, b) == 0 }
}

// Min returns the lesser value, or a when the values are order-equivalent.
func Min[T any](order Order[T], a, b T) T {
	if order(a, b) <= 0 {
		return a
	}
	return b
}

// Max returns the greater value, or a when the values are order-equivalent.
func Max[T any](order Order[T], a, b T) T {
	if order(a, b) >= 0 {
		return a
	}
	return b
}
