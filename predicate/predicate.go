package predicate

// Predicate decides whether a value satisfies a condition.
type Predicate[T any] func(T) bool
