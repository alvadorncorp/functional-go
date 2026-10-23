package immutable

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
	"github.com/alvadorncorp/gopackages/functional/internal/engine/wbtree"
	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// SortedMap is an immutable key-value collection ordered by a caller-supplied
// comparator. Every write method returns a new SortedMap; the receiver is
// never changed. Unlike Map, iteration order is specified: All, Keys, and
// Values yield entries in ascending comparator order.
//
// The comparator cmp(a, b) returns a negative, zero, or positive result and
// must be a strict weak order that stays deterministic for the map's
// lifetime. Keys for which cmp returns 0 are the same key: Set replaces both
// the stored key and its value. Every derived version, including one emptied
// by Delete, keeps the comparator it was constructed with. A comparator panic
// propagates to the caller.
//
// Zero value: unlike Map and Set, the zero SortedMap has no comparator,
// because K has no default order. It reads as empty and Delete on it is a
// no-op, but Set and Builder() panic. Construct a SortedMap with
// EmptySortedMap or CollectSortedMap; a zero-safe getter must return
// EmptySortedMap(cmp).
//
// Representation: SortedMap holds an unexported wbtree.Root. The blank
// [0]func() field keeps SortedMap deliberately not ==-comparable regardless
// of the engine's representation, which the public contract forbids exposing.
type SortedMap[K, V any] struct {
	_    [0]func()
	root wbtree.Root[K, V]
}

// EmptySortedMap returns an empty SortedMap ordered by cmp. It panics with
// "immutable: nil comparator" when cmp is nil.
func EmptySortedMap[K, V any](cmp func(a, b K) int) SortedMap[K, V] {
	if cmp == nil {
		panic(panicNilComparator)
	}
	return SortedMap[K, V]{root: wbtree.New[K, V](cmp)}
}

// CollectSortedMap drains seq into a new SortedMap ordered by cmp. When seq
// yields equivalent keys more than once, the last key and the last value
// win. It panics with "immutable: nil comparator" when cmp is nil.
func CollectSortedMap[K, V any](cmp func(a, b K) int, seq iter.Seq2[K, V]) SortedMap[K, V] {
	b := EmptySortedMap[K, V](cmp).Builder()
	for k, v := range seq {
		b.Set(k, v)
	}
	return b.Build()
}

// writableRoot returns m's root, panicking with an actionable message when m
// is the zero SortedMap, so no "wbtree: ..." panic reaches the caller.
func (m SortedMap[K, V]) writableRoot() wbtree.Root[K, V] {
	if !m.root.HasComparator() {
		panic(panicZeroSortedMap)
	}
	return m.root
}

// Len returns the number of entries in m.
func (m SortedMap[K, V]) Len() int { return m.root.Len() }

// IsEmpty reports whether m holds no entries. The zero SortedMap is empty.
func (m SortedMap[K, V]) IsEmpty() bool { return m.root.Len() == 0 }

// Get returns the value stored for a key equivalent to k and true, or the
// zero value of V and false when no such key is present.
func (m SortedMap[K, V]) Get(k K) (V, bool) {
	return m.root.Get(k)
}

// Has reports whether a key equivalent to k is present in m.
func (m SortedMap[K, V]) Has(k K) bool {
	_, ok := m.root.Get(k)
	return ok
}

// Set returns a new SortedMap with k mapped to v. An equivalent stored key is
// replaced by k, and its value by v. m is unchanged. Set panics on the zero
// SortedMap.
func (m SortedMap[K, V]) Set(k K, v V) SortedMap[K, V] {
	return SortedMap[K, V]{root: m.writableRoot().Set(k, v)}
}

// Delete returns a new SortedMap without k. m is unchanged. Delete on the
// zero SortedMap is a no-op.
func (m SortedMap[K, V]) Delete(k K) SortedMap[K, V] {
	return SortedMap[K, V]{root: m.root.Delete(k)}
}

// All returns an iterator over m's (key, value) pairs in ascending
// comparator order.
func (m SortedMap[K, V]) All() iter.Seq2[K, V] { return m.root.All() }

// Keys returns an iterator over m's keys in ascending comparator order.
func (m SortedMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.root.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator over m's values in ascending order of their
// keys.
func (m SortedMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m.root.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// Equal reports whether m and other hold the same number of entries and
// every key in m is present in other with an equal value under
// functional/internal/eq's value equality. Keys are looked up with other's
// comparator; even a shared version compares its values.
func (m SortedMap[K, V]) Equal(other SortedMap[K, V]) bool {
	return m.EqualFunc(other, eq.For[V]())
}

// EqualFunc reports whether m and other hold the same number of entries and
// every key in m is present in other with a value for which equal returns
// true. Keys present only in other are rejected by the length check.
// Comparison stops at the first missing key or false result from equal.
func (m SortedMap[K, V]) EqualFunc(other SortedMap[K, V], equal func(V, V) bool) bool {
	if m.root.Len() != other.root.Len() {
		return false
	}
	for k, v := range m.root.All() {
		ov, ok := other.root.Get(k)
		if !ok || !equal(v, ov) {
			return false
		}
	}
	return true
}

// Builder returns a SortedMapBuilder seeded from m's root (same node graph;
// later exclusive edits copy-on-write) and comparator. m is not affected by
// any later operation on the builder. Builder panics on the zero SortedMap.
func (m SortedMap[K, V]) Builder() *SortedMapBuilder[K, V] {
	return &SortedMapBuilder[K, V]{
		root: m.writableRoot(),
		win:  transient.Open(),
	}
}

// SortedMapBuilder accumulates writes for SortedMap under an exclusive
// transient owner. It is MUTABLE and is NOT safe for concurrent use.
//
// Set/Delete edit owned tree nodes in place and copy shared ones. Build
// freezes the window (O(1) owner invalidation — no tree walk), returns a
// SortedMap with the current root, and permanently disables the builder. A
// comparator panic during Set or Delete leaves the builder's prior contents
// intact.
//
// After Build runs, every method on the builder — including Build itself —
// panics with "immutable: builder used after Build".
type SortedMapBuilder[K, V any] struct {
	root wbtree.Root[K, V]
	win  *transient.Window
}

// Set maps k to v under the builder's exclusive owner, replacing an
// equivalent stored key and its value, and returns b for chaining.
func (b *SortedMapBuilder[K, V]) Set(k K, v V) *SortedMapBuilder[K, V] {
	checkBuilderWindow(b.win)
	b.root = b.root.TransientSet(b.win.Owner(), k, v)
	return b
}

// Delete removes k under the builder's exclusive owner and returns b for
// chaining.
func (b *SortedMapBuilder[K, V]) Delete(k K) *SortedMapBuilder[K, V] {
	checkBuilderWindow(b.win)
	b.root = b.root.TransientDelete(b.win.Owner(), k)
	return b
}

// Len returns the number of entries currently held by the builder.
func (b *SortedMapBuilder[K, V]) Len() int {
	checkBuilderWindow(b.win)
	return b.root.Len()
}

// Get returns the value stored in the builder for a key equivalent to k and
// true, or the zero value of V and false.
func (b *SortedMapBuilder[K, V]) Get(k K) (V, bool) {
	checkBuilderWindow(b.win)
	return b.root.Get(k)
}

// Has reports whether a key equivalent to k is present in the builder.
func (b *SortedMapBuilder[K, V]) Has(k K) bool {
	checkBuilderWindow(b.win)
	_, ok := b.root.Get(k)
	return ok
}

// Build returns an immutable SortedMap holding the builder's current root and
// permanently disables the builder: every method, including a repeated
// Build, panics with "immutable: builder used after Build" from this point
// on. Freeze invalidates the owner token in O(1); nodes may retain stale
// inactive stamps.
func (b *SortedMapBuilder[K, V]) Build() SortedMap[K, V] {
	freezeBuilderWindow(&b.win)
	root := b.root
	b.root = wbtree.Root[K, V]{}
	return SortedMap[K, V]{root: root}
}
