package function

// Tap returns a function that calls fn with its input and returns the input unchanged.
func Tap[T any](fn func(T)) func(T) T {
	return func(value T) T {
		fn(value)
		return value
	}
}
