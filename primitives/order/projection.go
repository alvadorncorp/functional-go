package order

// By orders T values by projecting them to keys ordered by keyOrder.
func By[T, K any](projection func(T) K, keyOrder Order[K]) Order[T] {
	return func(a, b T) int { return keyOrder(projection(a), projection(b)) }
}
