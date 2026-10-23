package function

import "slices"

// Transform represents a function that transforms a value while
// preserving its type.
//
//	T -> T
//
// Mathematically, such a function is an endomorphism. The name
// Transform is used instead of Endo to make the API more explicit
// and idiomatic for Go code.
//
// Transforms are closed under composition: composing multiple
// Transform[T] values produces another Transform[T].
type Transform[T any] func(T) T

// PipeTransforms composes transforms left to right.
//
//	PipeTransforms(f, g, h)(x) == h(g(f(x)))
//
// With no transforms, PipeTransforms returns the identity transform.
// The variadic slice is copied at construction; later changes to the
// caller's slice do not affect the returned transform.
func PipeTransforms[T any](transforms ...Transform[T]) Transform[T] {
	transforms = slices.Clone(transforms)
	return func(value T) T {
		for _, transform := range transforms {
			value = transform(value)
		}
		return value
	}
}

// ComposeTransforms composes transforms right to left.
//
//	ComposeTransforms(f, g, h)(x) == f(g(h(x)))
//
// With no transforms, ComposeTransforms returns the identity transform.
// Direction matches Compose (outermost first in the argument list).
// The variadic slice is copied at construction; later changes to the
// caller's slice do not affect the returned transform.
func ComposeTransforms[T any](transforms ...Transform[T]) Transform[T] {
	transforms = slices.Clone(transforms)
	return func(value T) T {
		for i := len(transforms) - 1; i >= 0; i-- {
			value = transforms[i](value)
		}
		return value
	}
}
