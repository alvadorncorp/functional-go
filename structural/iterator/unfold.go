package iterator

import "iter"

func Iterate[T any](seed T, fn func(T) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		value := seed
		for yield(value) {
			value = fn(value)
		}
	}
}

func Unfold[S, T any](state S, fn func(S) (value T, next S, ok bool)) iter.Seq[T] {
	return func(yield func(T) bool) {
		current := state
		for {
			value, next, ok := fn(current)
			if !ok {
				return
			}
			if !yield(value) {
				return
			}
			current = next
		}
	}
}
