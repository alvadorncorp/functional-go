package persistent

import (
	"iter"
	"slices"

	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// List is an immutable, ordered collection of T. Every write method returns
// a new List; the receiver is never changed. See the package doc for the
// full contract, including why List is not ==-comparable and why documented
// complexity is an upper bound rather than today's constant.
type List[T any] struct {
	items []T
}

// NewList returns a List holding a copy of items, in order. The caller's
// slice is never retained: later mutation of items through the caller's own
// reference does not reach the returned List.
func NewList[T any](items ...T) List[T] {
	return List[T]{items: slices.Clone(items)}
}

// CollectList drains seq into a new List, in iteration order.
func CollectList[T any](seq iter.Seq[T]) List[T] {
	return List[T]{items: slices.Collect(seq)}
}

// Len returns the number of elements in l.
func (l List[T]) Len() int { return len(l.items) }

// At returns the element at index i. It panics with
// "persistent: index out of range" when i is outside [0, Len).
func (l List[T]) At(i int) T {
	if i < 0 || i >= len(l.items) {
		panic(panicIndexOutOfRange)
	}
	return l.items[i]
}

// Append returns a new List with items added to the end, in order. l is
// unchanged.
//
// A single Append is a full copy of the receiver: O(n) with today's engine.
// Calling Append n times in a loop is therefore O(n^2); use Builder for
// that case instead.
func (l List[T]) Append(items ...T) List[T] {
	next := make([]T, len(l.items)+len(items))
	copy(next, l.items)
	copy(next[len(l.items):], items)
	return List[T]{items: next}
}

// Set returns a new List with the element at index i replaced by v. l is
// unchanged. It panics with "persistent: index out of range" when i is
// outside [0, Len).
func (l List[T]) Set(i int, v T) List[T] {
	if i < 0 || i >= len(l.items) {
		panic(panicIndexOutOfRange)
	}
	next := slices.Clone(l.items)
	next[i] = v
	return List[T]{items: next}
}

// Delete returns a new List without the element at index i; every later
// element shifts down by one. l is unchanged. It panics with
// "persistent: index out of range" when i is outside [0, Len).
func (l List[T]) Delete(i int) List[T] {
	if i < 0 || i >= len(l.items) {
		panic(panicIndexOutOfRange)
	}
	next := make([]T, len(l.items)-1)
	copy(next, l.items[:i])
	copy(next[i:], l.items[i+1:])
	return List[T]{items: next}
}

// Values returns an iterator over l's elements in index order.
func (l List[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range l.items {
			if !yield(v) {
				return
			}
		}
	}
}

// All returns an iterator over l's (index, element) pairs in index order.
func (l List[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range l.items {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Items returns a new slice holding a copy of l's elements, in order. The
// caller owns the result: mutating it never reaches l, and later writes to l
// never reach a slice returned earlier.
func (l List[T]) Items() []T { return slices.Clone(l.items) }

// Equal reports whether l and other hold the same elements in the same
// order, comparing each pair with functional/internal/eq.Equal.
func (l List[T]) Equal(other List[T]) bool {
	if len(l.items) != len(other.items) {
		return false
	}
	for i := range l.items {
		if !eq.Equal(l.items[i], other.items[i]) {
			return false
		}
	}
	return true
}

// Builder returns a ListBuilder seeded with a copy of l's elements. l is not
// affected by any later operation on the builder.
func (l List[T]) Builder() *ListBuilder[T] {
	return &ListBuilder[T]{items: slices.Clone(l.items)}
}

// ListBuilder accumulates writes for List in amortized linear time. It is
// MUTABLE and is NOT safe for concurrent use.
//
// Build transfers ownership of the builder's internal buffer to the List it
// returns. After Build runs, every method on the builder — including Build
// itself — panics with "persistent: builder used after Build".
type ListBuilder[T any] struct {
	items []T
	built bool
}

func (b *ListBuilder[T]) checkNotBuilt() {
	if b.built {
		panic(panicBuilderUsedAfterBuild)
	}
}

// Append adds items to the end of the builder's buffer, in order, and
// returns b for chaining.
func (b *ListBuilder[T]) Append(items ...T) *ListBuilder[T] {
	b.checkNotBuilt()
	b.items = append(b.items, items...)
	return b
}

// Set replaces the element at index i with v and returns b for chaining. It
// panics with "persistent: index out of range" when i is outside [0, Len).
func (b *ListBuilder[T]) Set(i int, v T) *ListBuilder[T] {
	b.checkNotBuilt()
	if i < 0 || i >= len(b.items) {
		panic(panicIndexOutOfRange)
	}
	b.items[i] = v
	return b
}

// Delete removes the element at index i, shifting every later element down
// by one, and returns b for chaining. It panics with
// "persistent: index out of range" when i is outside [0, Len).
func (b *ListBuilder[T]) Delete(i int) *ListBuilder[T] {
	b.checkNotBuilt()
	if i < 0 || i >= len(b.items) {
		panic(panicIndexOutOfRange)
	}
	b.items = append(b.items[:i], b.items[i+1:]...)
	return b
}

// Len returns the number of elements currently buffered.
func (b *ListBuilder[T]) Len() int {
	b.checkNotBuilt()
	return len(b.items)
}

// Build returns an immutable List holding the builder's buffered elements
// and permanently disables the builder: every method, including a repeated
// Build, panics with "persistent: builder used after Build" from this point
// on.
func (b *ListBuilder[T]) Build() List[T] {
	b.checkNotBuilt()
	b.built = true
	items := b.items
	b.items = nil
	return List[T]{items: items}
}
