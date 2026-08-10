package order

// Then composes primary and secondary lexicographically.
func Then[T any](primary, secondary Order[T]) Order[T] {
	return func(a, b T) int {
		if result := primary(a, b); result != 0 {
			return result
		}
		return secondary(a, b)
	}
}
