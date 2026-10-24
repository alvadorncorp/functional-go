package immutable

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
// Set is a valid empty set for reads (Len is 0, Has misses). The first write
// (Add/Remove) or Builder() call materializes a proper Empty-seeded root so
// the lineage never freezes a zero maphash.Seed. Constructed Sets
// (SetOf, CollectSet, Builder.Build, and every Add/Remove result) already
// hold an Empty-seeded root when they first materialize entries.
//
// The blank [0]func() field keeps Set deliberately not ==-comparable:
// hamt.Root's fields are all comparable, so without this marker Set would
// compare by representation identity, which the public contract forbids.
type Set[T comparable] struct {
	_    [0]func()
	root hamt.Root[T, struct{}]
}

// materialize returns s's root, replacing an unseeded zero hamt.Root with
// hamt.Empty so the first write starts a properly seeded lineage.
func (s Set[T]) materialize() hamt.Root[T, struct{}] {
	var zero hamt.Root[T, struct{}]
	if s.root == zero {
		return hamt.Empty[T, struct{}]()
	}
	return s.root
}

// SetOf returns a Set holding items, deduplicated. When items repeat, the
// last equal item is the stored representative. Empty input returns the zero
// Set without a hash seed.
func SetOf[T comparable](items ...T) Set[T] {
	return (Set[T]{}).Add(items...)
}

// CollectSet drains seq into a new Set, deduplicated. When seq repeats an
// element, the last equal element is the stored representative.
func CollectSet[T comparable](seq iter.Seq[T]) Set[T] {
	var result Set[T]
	var b *SetBuilder[T]
	for v := range seq {
		switch {
		case b != nil:
			b.Add(v)
		case result.IsEmpty():
			result = result.Add(v)
		default:
			b = result.Builder().Add(v)
		}
	}
	if b == nil {
		return result
	}
	return b.Build()
}

// Len returns the number of elements in s.
func (s Set[T]) Len() int { return s.root.Len() }

// IsEmpty reports whether s holds no elements. The zero Set is empty.
func (s Set[T]) IsEmpty() bool { return s.root.Len() == 0 }

// Has reports whether v is a member of s.
func (s Set[T]) Has(v T) bool {
	_, ok := s.root.Get(v)
	return ok
}

// Add returns a new Set with items added. An item equal to one already held
// replaces the stored representative, which is observable for values that
// compare equal but are distinguishable (+0 and -0, strings sharing
// different backing data). s is unchanged.
func (s Set[T]) Add(items ...T) Set[T] {
	switch len(items) {
	case 0:
		return s
	case 1:
		return Set[T]{root: s.materialize().Set(items[0], struct{}{})}
	default:
		return s.Builder().Add(items...).Build()
	}
}

// Remove returns a new Set without items; items not present are a no-op. s
// is unchanged.
func (s Set[T]) Remove(items ...T) Set[T] {
	switch len(items) {
	case 0:
		return s
	case 1:
		return Set[T]{root: s.materialize().Delete(items[0])}
	default:
		return s.Builder().Remove(items...).Build()
	}
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
// eq.Equal comparison is needed. A shared root is always equal to itself,
// even when an element is NaN.
func (s Set[T]) Equal(other Set[T]) bool {
	if s.root == other.root {
		return true
	}
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

// Union returns a new Set holding every element of s or other. s and other
// are unchanged. Cost is O(min(m,n)·log max(m,n)) for operand sizes m and n:
// the smaller operand is walked and probed against the larger. Which
// operand's structure the result shares, and which operand's hash-seed
// lineage it continues, is UNSPECIFIED.
func (s Set[T]) Union(other Set[T]) Set[T] {
	if s.root == other.root {
		return s
	}
	base, walk := s, other
	if other.Len() > s.Len() {
		base, walk = other, s
	}
	return addMissing(base, walk)
}

// Intersection returns a new Set holding every element of s that other also
// holds. s and other are unchanged. Cost is O(min(m,n)·log max(m,n)) for
// operand sizes m and n: the smaller operand is walked and probed against
// the larger. Which operand's structure the result shares, and which
// operand's hash-seed lineage it continues, is UNSPECIFIED.
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	if s.root == other.root {
		return s
	}
	base, probe := s, other
	if other.Len() < s.Len() {
		base, probe = other, s
	}
	return removeIf(base, base, func(v T) bool { return !probe.Has(v) })
}

// Difference returns a new Set holding every element of s that other does
// not hold (s minus other). s and other are unchanged. Cost is
// O(min(m,n)·log max(m,n)) for operand sizes m and n: the smaller operand is
// walked and probed against the larger. Which operand's structure the result
// shares, and which operand's hash-seed lineage it continues, is UNSPECIFIED.
func (s Set[T]) Difference(other Set[T]) Set[T] {
	if s.root == other.root {
		return Set[T]{}
	}
	if s.Len() <= other.Len() {
		return removeIf(s, s, other.Has)
	}
	return removeIf(s, other, s.Has)
}

// addMissing returns base plus every element of walk that base lacks. A
// SetBuilder opens on base only at the first addition, so when walk adds
// nothing (including an empty walk) base is returned untouched.
func addMissing[T comparable](base, walk Set[T]) Set[T] {
	var b *SetBuilder[T]
	for v := range walk.root.All() {
		if base.Has(v) {
			continue
		}
		if b == nil {
			b = base.Builder()
		}
		b.Add(v)
	}
	if b == nil {
		return base
	}
	return b.Build()
}

// removeIf returns base minus every element v of walk for which drop(v)
// holds; walk is base itself or a set whose matching elements are all in
// base. A SetBuilder opens on base only at the first removal, so when
// nothing is dropped (including an empty walk) base is returned untouched.
// Walking base while the builder edits it is safe: the builder's fresh owner
// copy-on-writes every node of base before changing it.
func removeIf[T comparable](base, walk Set[T], drop func(T) bool) Set[T] {
	var b *SetBuilder[T]
	for v := range walk.root.All() {
		if !drop(v) {
			continue
		}
		if b == nil {
			b = base.Builder()
		}
		b.Remove(v)
	}
	if b == nil {
		return base
	}
	return b.Build()
}

// Builder returns a SetBuilder seeded from s's root (same node graph; later
// exclusive edits copy-on-write). s is not affected by any later operation
// on the builder. A zero Set materializes an Empty-seeded root first.
func (s Set[T]) Builder() *SetBuilder[T] {
	return &SetBuilder[T]{
		root: s.materialize(),
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
// panics with "immutable: builder used after Build".
type SetBuilder[T comparable] struct {
	root hamt.Root[T, struct{}]
	win  *transient.Window
}

func (b *SetBuilder[T]) checkNotBuilt() {
	checkBuilderWindow(b.win)
}

// Add adds items under the builder's exclusive owner, deduplicated, and
// returns b for chaining. As with Set.Add, an item equal to one already held
// replaces the stored representative.
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

// Has reports whether v is a member of the builder. After Build, Has panics
// with "immutable: builder used after Build".
func (b *SetBuilder[T]) Has(v T) bool {
	b.checkNotBuilt()
	_, ok := b.root.Get(v)
	return ok
}

// Build returns an immutable Set holding the builder's current root and
// permanently disables the builder: every method, including a repeated
// Build, panics with "immutable: builder used after Build" from this point
// on. Freeze invalidates the owner token in O(1); nodes may retain stale
// inactive stamps.
func (b *SetBuilder[T]) Build() Set[T] {
	freezeBuilderWindow(&b.win)
	root := b.root
	b.root = hamt.Root[T, struct{}]{}
	return Set[T]{root: root}
}
