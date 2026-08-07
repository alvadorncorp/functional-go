package persistent

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
)

// Set is an immutable collection of distinct, comparable values. Every write
// method returns a new Set; the receiver is never changed. See the package
// doc for the full contract, including why Set is not ==-comparable, why
// iteration order is unspecified, and why documented complexity is an upper
// bound rather than today's constant.
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet returns a Set holding items, deduplicated.
func NewSet[T comparable](items ...T) Set[T] {
	next := make(map[T]struct{}, len(items))
	for _, v := range items {
		next[v] = struct{}{}
	}
	return Set[T]{items: next}
}

// CollectSet drains seq into a new Set, deduplicated.
func CollectSet[T comparable](seq iter.Seq[T]) Set[T] {
	next := map[T]struct{}{}
	for v := range seq {
		next[v] = struct{}{}
	}
	return Set[T]{items: next}
}

// Len returns the number of elements in s.
func (s Set[T]) Len() int { return len(s.items) }

// Has reports whether v is a member of s.
func (s Set[T]) Has(v T) bool {
	_, ok := s.items[v]
	return ok
}

// Add returns a new Set with items added; items already present are a
// no-op. s is unchanged.
func (s Set[T]) Add(items ...T) Set[T] {
	next := cloneMap(s.items, len(items))
	for _, v := range items {
		next[v] = struct{}{}
	}
	return Set[T]{items: next}
}

// Remove returns a new Set without items; items not present are a no-op. s
// is unchanged.
func (s Set[T]) Remove(items ...T) Set[T] {
	next := cloneMap(s.items, 0)
	for _, v := range items {
		delete(next, v)
	}
	return Set[T]{items: next}
}

// Values returns an iterator over s's elements. Iteration order is
// UNSPECIFIED and may change between releases.
func (s Set[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s.items {
			if !yield(v) {
				return
			}
		}
	}
}

// Items returns a new slice holding s's elements. Iteration order is
// UNSPECIFIED and may change between releases. The caller owns the result:
// mutating it never reaches s.
func (s Set[T]) Items() []T {
	out := make([]T, 0, len(s.items))
	for v := range s.items {
		out = append(out, v)
	}
	return out
}

// Equal reports whether s and other hold the same elements. Because T is
// comparable, a value's membership alone decides equality; no per-element
// eq.Equal comparison is needed.
func (s Set[T]) Equal(other Set[T]) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for v := range s.items {
		if _, ok := other.items[v]; !ok {
			return false
		}
	}
	return true
}

// Builder returns a SetBuilder seeded with a copy of s's elements. s is not
// affected by any later operation on the builder.
func (s Set[T]) Builder() *SetBuilder[T] {
	return &SetBuilder[T]{
		items: cloneMap(s.items, 0),
		win:   transient.Open(),
	}
}

// SetBuilder accumulates writes for Set in amortized linear time. It is
// MUTABLE and is NOT safe for concurrent use.
//
// Build transfers ownership of the builder's internal buffer to the Set it
// returns. After Build runs, every method on the builder — including Build
// itself — panics with "persistent: builder used after Build".
type SetBuilder[T comparable] struct {
	items map[T]struct{}
	win   *transient.Window
}

func (b *SetBuilder[T]) checkNotBuilt() {
	checkBuilderWindow(b.win)
}

// Add adds items to the builder's buffer, deduplicated, and returns b for
// chaining.
func (b *SetBuilder[T]) Add(items ...T) *SetBuilder[T] {
	b.checkNotBuilt()
	for _, v := range items {
		b.items[v] = struct{}{}
	}
	return b
}

// Remove removes items from the builder's buffer and returns b for
// chaining.
func (b *SetBuilder[T]) Remove(items ...T) *SetBuilder[T] {
	b.checkNotBuilt()
	for _, v := range items {
		delete(b.items, v)
	}
	return b
}

// Len returns the number of elements currently buffered.
func (b *SetBuilder[T]) Len() int {
	b.checkNotBuilt()
	return len(b.items)
}

// Build returns an immutable Set holding the builder's buffered elements
// and permanently disables the builder: every method, including a repeated
// Build, panics with "persistent: builder used after Build" from this point
// on.
func (b *SetBuilder[T]) Build() Set[T] {
	freezeBuilderWindow(&b.win)
	items := b.items
	b.items = nil
	return Set[T]{items: items}
}
