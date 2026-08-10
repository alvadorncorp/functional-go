package iterator

import "iter"

func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	if n <= 0 {
		return Empty[T]()
	}
	return func(yield func(T) bool) {
		remaining := n
		for value := range seq {
			remaining--
			if !yield(value) || remaining == 0 {
				return
			}
		}
	}
}

func Drop[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	if n <= 0 {
		return seq
	}
	return func(yield func(T) bool) {
		skipped := 0
		for value := range seq {
			if skipped < n {
				skipped++
				continue
			}
			if !yield(value) {
				return
			}
		}
	}
}

func TakeWhile[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for value := range seq {
			if !predicate(value) || !yield(value) {
				return
			}
		}
	}
}

func DropWhile[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		dropping := true
		for value := range seq {
			if dropping && predicate(value) {
				continue
			}
			dropping = false
			if !yield(value) {
				return
			}
		}
	}
}
