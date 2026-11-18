package iterator

import (
	"iter"

	"go.alvadorn.org/functional/structural/tuple"
)

func Zip[A, B any](left iter.Seq[A], right iter.Seq[B]) iter.Seq[tuple.Tuple2[A, B]] {
	return func(yield func(tuple.Tuple2[A, B]) bool) {
		nextLeft, stopLeft := iter.Pull(left)
		defer stopLeft()
		nextRight, stopRight := iter.Pull(right)
		defer stopRight()
		for {
			leftValue, leftOK := nextLeft()
			if !leftOK {
				return
			}
			rightValue, rightOK := nextRight()
			if !rightOK {
				return
			}
			if !yield(tuple.Of2(leftValue, rightValue)) {
				return
			}
		}
	}
}

func Enumerate[T any](seq iter.Seq[T]) iter.Seq[tuple.Tuple2[int, T]] {
	return func(yield func(tuple.Tuple2[int, T]) bool) {
		index := 0
		for value := range seq {
			if !yield(tuple.Of2(index, value)) {
				return
			}
			index++
		}
	}
}
