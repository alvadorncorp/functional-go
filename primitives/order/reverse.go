package order

// Reverse returns the opposite order by swapping arguments.
func Reverse[T any](order Order[T]) Order[T] {
	return func(a, b T) int { return order(b, a) }
}
