package frozen

import (
	"iter"
	"slices"

	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// outOfRangePanicMessage is what At panics with when i is outside [0, Len).
const outOfRangePanicMessage = "frozen: index out of range"

// Slice is an immutable snapshot of a []T. The zero Slice is a valid, empty
// Slice.
//
// Slice is not ==-comparable (it holds a slice header). A value object with a
// Slice field must compare it with Equal, not ==.
type Slice[T any] struct {
	items []T
}

// NewSlice copies items into a Slice. Passing an existing slice with xs...
// still copies: later writes to xs cannot reach the Slice.
func NewSlice[T any](items ...T) Slice[T] {
	return Slice[T]{items: slices.Clone(items)}
}

// CollectSlice copies the elements produced by seq into a Slice.
func CollectSlice[T any](seq iter.Seq[T]) Slice[T] {
	return Slice[T]{items: slices.Collect(seq)}
}

// Len returns the number of elements in s.
func (s Slice[T]) Len() int { return len(s.items) }

// At returns the element at index i. At panics with
// "frozen: index out of range" for i outside [0, Len).
func (s Slice[T]) At(i int) T {
	if i < 0 || i >= len(s.items) {
		panic(outOfRangePanicMessage)
	}
	return s.items[i]
}

// Values iterates the elements without allocating. Prefer it over Items.
func (s Slice[T]) Values() iter.Seq[T] {
	return slices.Values(s.items)
}

// All iterates index/value pairs, mirroring slices.All.
func (s Slice[T]) All() iter.Seq2[int, T] {
	return slices.All(s.items)
}

// Items returns a fresh mutable copy of the elements. Every call allocates.
func (s Slice[T]) Items() []T {
	return slices.Clone(s.items)
}

// Equal reports whether s and other hold the same elements in the same
// order, comparing element-wise via functional/internal/eq.Equal.
func (s Slice[T]) Equal(other Slice[T]) bool {
	return SliceEqualFunc(s, other, eq.Equal[T])
}
