package iterator

import (
	"iter"
	"slices"
)

func Empty[T any]() iter.Seq[T] { return func(func(T) bool) {} }

func Once[T any](value T) iter.Seq[T] {
	return func(yield func(T) bool) { yield(value) }
}

// FromSlice returns a lazy live view of values; it does not copy the slice.
//
// Deprecated: Use slices.Values instead.
func FromSlice[T any](values []T) iter.Seq[T] { return slices.Values(values) }

// FromValues returns a lazy live view of values; it does not copy the slice.
//
// Deprecated: Use slices.Values with a slice instead.
func FromValues[T any](values ...T) iter.Seq[T] { return slices.Values(values) }

func Repeat[T any](value T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for yield(value) {
		}
	}
}
