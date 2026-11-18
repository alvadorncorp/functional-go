package fold

import (
	"go.alvadorn.org/functional/values/nonempty"
	"go.alvadorn.org/functional/values/optional"
)

// Reduce combines values from left to right, using the first value as the
// initial state. Empty values return None.
func Reduce[T any](values []T, fn func(T, T) T) optional.Optional[T] {
	if len(values) == 0 {
		return optional.None[T]()
	}
	state := values[0]
	for _, value := range values[1:] {
		state = fn(state, value)
	}
	return optional.Some(state)
}

// ReduceNonEmpty combines a NonEmpty from left to right without an absence result.
func ReduceNonEmpty[T any](values nonempty.NonEmpty[T], fn func(T, T) T) T {
	state := values.Head()
	for index := 1; index < values.Len(); index++ {
		state = fn(state, values.At(index))
	}
	return state
}
