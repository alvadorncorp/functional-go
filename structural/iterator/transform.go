package iterator

import "iter"

func Map[A, B any](seq iter.Seq[A], fn func(A) B) iter.Seq[B] {
	return func(yield func(B) bool) {
		for value := range seq {
			if !yield(fn(value)) {
				return
			}
		}
	}
}

func Filter[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for value := range seq {
			if predicate(value) && !yield(value) {
				return
			}
		}
	}
}

func FilterMap[A, B any](seq iter.Seq[A], fn func(A) (B, bool)) iter.Seq[B] {
	return func(yield func(B) bool) {
		for value := range seq {
			mapped, ok := fn(value)
			if ok && !yield(mapped) {
				return
			}
		}
	}
}

func FlatMap[A, B any](seq iter.Seq[A], fn func(A) iter.Seq[B]) iter.Seq[B] {
	return func(yield func(B) bool) {
		for outer := range seq {
			for inner := range fn(outer) {
				if !yield(inner) {
					return
				}
			}
		}
	}
}

func Flatten[T any](seq iter.Seq[iter.Seq[T]]) iter.Seq[T] {
	return FlatMap(seq, func(inner iter.Seq[T]) iter.Seq[T] { return inner })
}
