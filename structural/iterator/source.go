package iterator

import (
	"iter"
	"slices"
)

func Empty[T any]() iter.Seq[T] { return func(func(T) bool) {} }

func Once[T any](value T) iter.Seq[T] {
	return func(yield func(T) bool) { yield(value) }
}

func FromSlice[T any](values []T) iter.Seq[T] { return slices.Values(values) }

func FromValues[T any](values ...T) iter.Seq[T] { return slices.Values(values) }

func Repeat[T any](value T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for yield(value) {
		}
	}
}
