package persistent

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
	"github.com/alvadorncorp/gopackages/functional/internal/engine/hamt"
)

// Set is an immutable collection of distinct, comparable values. Every write
// method returns a new Set; the receiver is never changed. See the package
// doc for the full contract, including why Set is not ==-comparable, why
// iteration order is unspecified, and why documented complexity is an upper
// bound rather than today's constant.
//
// Representation: Set holds an unexported hamt.Root[T, struct{}]. The zero
// Set is a valid empty set — it relies on the zero hamt.Root (nil node, zero
// seed, size 0): Len is 0, Has misses, and Add starts a lineage from that
// seed. Constructed Sets (NewSet, CollectSet, Builder.Build, and every
// Add/Remove result) use a proper Empty-seeded root when they first
// materialize entries.
//
// The blank [0]func() field keeps Set deliberately not ==-comparable:
// hamt.Root's fields are all comparable, so without this marker Set would
// compare by representation identity, which the public contract forbids.
type Set[T comparable] struct {
	_    [0]func()
	root hamt.Root[T, struct{}]
}

// NewSet returns a Set holding items, deduplicated.
func NewSet[T comparable](items ...T) Set[T] {
	return Set[T]{root: valuesToRoot(items)}
}

// CollectSet drains seq into a new Set, deduplicated.
func CollectSet[T comparable](seq iter.Seq[T]) Set[T] {
	root := hamt.Empty[T, struct{}]()
	for v := range seq {
		root = root.Set(v, struct{}{})
	}
	return Set[T]{root: root}
}

// valuesToRoot inserts values into a fresh Empty-seeded hamt.Root.
func valuesToRoot[T comparable](values []T) hamt.Root[T, struct{}] {
	root := hamt.Empty[T, struct{}]()
	for _, v := range values {
		root = root.Set(v, struct{}{})
	}
	return root
}

// Len returns the number of elements in s.
func (s Set[T]) Len() int { return s.root.Len() }

// Has reports whether v is a member of s.
func (s Set[T]) Has(v T) bool {
	_, ok := s.root.Get(v)
	return ok
}

// Add returns a new Set with items added; items already present are a
// no-op. s is unchanged.
func (s Set[T]) Add(items ...T) Set[T] {
	root := s.root
	for _, v := range items {
		root = root.Set(v, struct{}{})
	}
	return Set[T]{root: root}
}

// Remove returns a new Set without items; items not present are a no-op. s
// is unchanged.
func (s Set[T]) Remove(items ...T) Set[T] {
	root := s.root
	for _, v := range items {
		root = root.Delete(v)
	}
	return Set[T]{root: root}
}

// Values returns an iterator over s's elements. Iteration order is
// UNSPECIFIED and may change between releases.
func (s Set[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s.root.All() {
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
	out := make([]T, 0, s.root.Len())
	for v := range s.root.All() {
		out = append(out, v)
	}
	return out
}

// Equal reports whether s and other hold the same elements. Because T is
// comparable, a value's membership alone decides equality; no per-element
// eq.Equal comparison is needed.
func (s Set[T]) Equal(other Set[T]) bool {
	if s.root.Len() != other.root.Len() {
		return false
	}
	for v := range s.root.All() {
		if _, ok := other.root.Get(v); !ok {
			return false
		}
	}
	return true
}

// Builder returns a SetBuilder seeded from s's root (same node graph; later
// exclusive edits copy-on-write). s is not affected by any later operation
// on the builder.
func (s Set[T]) Builder() *SetBuilder[T] {
	return &SetBuilder[T]{
		root: s.root,
		win:  transient.Open(),
	}
}

// SetBuilder accumulates writes for Set in amortized linear time via the
// HAMT transient path. It is MUTABLE and is NOT safe for concurrent use.
//
// The builder holds a hamt.Root and an open transient.Window. Add/Remove
// call engine transient ops under win.Owner(). Build freezes the window
// (O(1) owner invalidation — no tree walk), returns a Set with the current
// root, and permanently disables the builder.
//
// After Build runs, every method on the builder — including Build itself —
// panics with "persistent: builder used after Build".
type SetBuilder[T comparable] struct {
	root hamt.Root[T, struct{}]
	win  *transient.Window
}

func (b *SetBuilder[T]) checkNotBuilt() {
	checkBuilderWindow(b.win)
}

// Add adds items under the builder's exclusive owner, deduplicated, and
// returns b for chaining.
func (b *SetBuilder[T]) Add(items ...T) *SetBuilder[T] {
	b.checkNotBuilt()
	owner := b.win.Owner()
	for _, v := range items {
		b.root = b.root.TransientSet(owner, v, struct{}{})
	}
	return b
}

// Remove removes items under the builder's exclusive owner and returns b
// for chaining.
func (b *SetBuilder[T]) Remove(items ...T) *SetBuilder[T] {
	b.checkNotBuilt()
	owner := b.win.Owner()
	for _, v := range items {
		b.root = b.root.TransientDelete(owner, v)
	}
	return b
}

// Len returns the number of elements currently held by the builder.
func (b *SetBuilder[T]) Len() int {
	b.checkNotBuilt()
	return b.root.Len()
}

// Build returns an immutable Set holding the builder's current root and
// permanently disables the builder: every method, including a repeated
// Build, panics with "persistent: builder used after Build" from this point
// on. Freeze invalidates the owner token in O(1); nodes may retain stale
// inactive stamps.
func (b *SetBuilder[T]) Build() Set[T] {
	freezeBuilderWindow(&b.win)
	root := b.root
	b.root = hamt.Root[T, struct{}]{}
	return Set[T]{root: root}
}
