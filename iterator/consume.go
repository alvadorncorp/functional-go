package iterator

import (
	"iter"
	"slices"
)

func Collect[T any](seq iter.Seq[T]) []T { return slices.Collect(seq) }

func Fold[T, A any](seq iter.Seq[T], initial A, fn func(A, T) A) A {
	result := initial
	for value := range seq {
		result = fn(result, value)
	}
	return result
}

func Any[T any](seq iter.Seq[T], predicate func(T) bool) bool {
	for value := range seq {
		if predicate(value) {
			return true
		}
	}
	return false
}

func All[T any](seq iter.Seq[T], predicate func(T) bool) bool {
	for value := range seq {
		if !predicate(value) {
			return false
		}
	}
	return true
}

func Find[T any](seq iter.Seq[T], predicate func(T) bool) (T, bool) {
	for value := range seq {
		if predicate(value) {
			return value, true
		}
	}
	var zero T
	return zero, false
}

func First[T any](seq iter.Seq[T]) (T, bool) {
	for value := range seq {
		return value, true
	}
	var zero T
	return zero, false
}

func Last[T any](seq iter.Seq[T]) (T, bool) {
	var last T
	found := false
	for value := range seq {
		last, found = value, true
	}
	return last, found
}

func Count[T any](seq iter.Seq[T]) int {
	count := 0
	for range seq {
		count++
	}
	return count
}
