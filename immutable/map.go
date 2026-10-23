package immutable

import (
	"iter"
	"maps"

	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
	"github.com/alvadorncorp/gopackages/functional/internal/engine/hamt"
	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// Map is an immutable key-value collection. Every write method returns a
// new Map; the receiver is never changed. See the package doc for the full
// contract, including why Map is not ==-comparable, why iteration order is
// unspecified, and why documented complexity is an upper bound rather than
// a guarantee of today's constants.
//
// Representation: Map holds an unexported hamt.Root. The zero Map is a
// valid empty map for reads (Len is 0, Get misses). The first write
// (Set/Update/Delete) or Builder() call materializes a proper Empty-seeded
// root so the lineage never freezes a zero maphash.Seed. Constructed Maps
// (MapOf, CollectMap, Builder.Build, and every Set/Update/Delete result)
// already hold an Empty-seeded root when they first materialize entries.
//
// The blank [0]func() field keeps Map deliberately not ==-comparable:
// hamt.Root's fields are all comparable, so without this marker Map would
// compare by representation identity, which the public contract forbids.
type Map[K comparable, V any] struct {
	_    [0]func()
	root hamt.Root[K, V]
}

// materialize returns m's root, replacing an unseeded zero hamt.Root with
// hamt.Empty so the first write starts a properly seeded lineage.
func (m Map[K, V]) materialize() hamt.Root[K, V] {
	var zero hamt.Root[K, V]
	if m.root == zero {
		return hamt.Empty[K, V]()
	}
	return m.root
}

// MapOf returns a Map holding a copy of m. The caller's map is never
// retained: later mutation of m through the caller's own reference does not
// reach the returned Map.
func MapOf[K comparable, V any](m map[K]V) Map[K, V] {
	if len(m) <= 1 {
		var result Map[K, V]
		for k, v := range m {
			result = result.Set(k, v)
		}
		return result
	}
	b := (Map[K, V]{}).Builder()
	for k, v := range m {
		b.Set(k, v)
	}
	return b.Build()
}

// CollectMap drains seq into a new Map. When seq yields equivalent keys more
// than once, the last key and value win, matching maps.Collect.
func CollectMap[K comparable, V any](seq iter.Seq2[K, V]) Map[K, V] {
	var result Map[K, V]
	var b *MapBuilder[K, V]
	for k, v := range seq {
		switch {
		case b != nil:
			b.Set(k, v)
		case result.IsEmpty():
			result = result.Set(k, v)
		default:
			b = result.Builder().Set(k, v)
		}
	}
	if b == nil {
		return result
	}
	return b.Build()
}

// Len returns the number of entries in m.
func (m Map[K, V]) Len() int { return m.root.Len() }

// IsEmpty reports whether m holds no entries. The zero Map is empty.
func (m Map[K, V]) IsEmpty() bool { return m.root.Len() == 0 }

// Get returns the value stored for k and true, or the zero value of V and
// false when k is absent.
func (m Map[K, V]) Get(k K) (V, bool) {
	return m.root.Get(k)
}

// Has reports whether k is present in m.
func (m Map[K, V]) Has(k K) bool {
	_, ok := m.root.Get(k)
	return ok
}

// Set returns a new Map with k mapped to v. An equivalent stored key is
// replaced by k, and its value by v. m is unchanged.
func (m Map[K, V]) Set(k K, v V) Map[K, V] {
	return Map[K, V]{root: m.materialize().Set(k, v)}
}

// Update returns a new Map with k mapped to fn(old, present), where old and
// present are what Get(k) returns: the stored value and true, or the zero
// value of V and false. fn is called exactly once, before anything is
// written, and its result is always stored. m is unchanged. A nil fn is not
// special-cased; calling Update with it panics like any nil function call.
func (m Map[K, V]) Update(k K, fn func(old V, present bool) V) Map[K, V] {
	old, ok := m.root.Get(k)
	return m.Set(k, fn(old, ok))
}

// Delete returns a new Map without k. m is unchanged.
func (m Map[K, V]) Delete(k K) Map[K, V] {
	return Map[K, V]{root: m.materialize().Delete(k)}
}

// All returns an iterator over m's (key, value) pairs. Iteration order is
// UNSPECIFIED and may change between releases.
func (m Map[K, V]) All() iter.Seq2[K, V] { return m.root.All() }

// Keys returns an iterator over m's keys. Iteration order is UNSPECIFIED and
// may change between releases.
func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.root.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator over m's values. Iteration order is
// UNSPECIFIED and may change between releases.
func (m Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m.root.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// Items returns a new, non-nil map holding a copy of m's entries. The
// caller owns the result: mutating it never reaches m, and later writes to m
// never reach a map returned earlier.
func (m Map[K, V]) Items() map[K]V {
	out := make(map[K]V, m.root.Len())
	maps.Insert(out, m.root.All())
	return out
}

// Equal reports whether m and other hold the same set of keys, each mapped
// to an equal value under functional/internal/eq's value equality. A shared
// root is always equal to itself, even with NaN or non-reflexive Equal methods.
func (m Map[K, V]) Equal(other Map[K, V]) bool {
	if m.root == other.root {
		return true
	}
	if m.root.Len() != other.root.Len() {
		return false
	}
	equal := eq.For[V]()
	for k, v := range m.root.All() {
		ov, ok := other.root.Get(k)
		if !ok || !equal(v, ov) {
			return false
		}
	}
	return true
}

// EqualFunc reports whether m and other hold the same number of entries and
// every key in m is present in other with a value for which equal returns true.
// Keys present only in other are rejected by the length check. Comparison
// stops at the first missing key or false result from equal.
func (m Map[K, V]) EqualFunc(other Map[K, V], equal func(V, V) bool) bool {
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

// Builder returns a MapBuilder seeded from m's root (same node graph; later
// exclusive edits copy-on-write). m is not affected by any later operation
// on the builder. A zero Map materializes an Empty-seeded root first.
func (m Map[K, V]) Builder() *MapBuilder[K, V] {
	return &MapBuilder[K, V]{
		root: m.materialize(),
		win:  transient.Open(),
	}
}

// MapBuilder accumulates writes for Map in amortized linear time via the
// HAMT transient path. It is MUTABLE and is NOT safe for concurrent use.
//
// The builder holds a hamt.Root and an open transient.Window. Set/Delete
// call engine transient ops under win.Owner(). Build freezes the window
// (O(1) owner invalidation — no tree walk), returns a Map with the current
// root, and permanently disables the builder.
//
// After Build runs, every method on the builder — including Build itself —
// panics with "immutable: builder used after Build".
type MapBuilder[K comparable, V any] struct {
	root hamt.Root[K, V]
	win  *transient.Window
}

func (b *MapBuilder[K, V]) checkNotBuilt() {
	checkBuilderWindow(b.win)
}

// Set maps k to v under the builder's exclusive owner, replacing an equivalent
// stored key by k and its value by v, and returns b for chaining.
func (b *MapBuilder[K, V]) Set(k K, v V) *MapBuilder[K, V] {
	b.checkNotBuilt()
	b.root = b.root.TransientSet(b.win.Owner(), k, v)
	return b
}

// Delete removes k under the builder's exclusive owner and returns b for
// chaining.
func (b *MapBuilder[K, V]) Delete(k K) *MapBuilder[K, V] {
	b.checkNotBuilt()
	b.root = b.root.TransientDelete(b.win.Owner(), k)
	return b
}

// Len returns the number of entries currently held by the builder.
func (b *MapBuilder[K, V]) Len() int {
	b.checkNotBuilt()
	return b.root.Len()
}

// Get returns the value stored for k in the builder and true, or the zero
// value of V and false when k is absent. After Build, Get panics with
// "immutable: builder used after Build".
func (b *MapBuilder[K, V]) Get(k K) (V, bool) {
	b.checkNotBuilt()
	return b.root.Get(k)
}

// Has reports whether k is present in the builder. After Build, Has panics
// with "immutable: builder used after Build".
func (b *MapBuilder[K, V]) Has(k K) bool {
	b.checkNotBuilt()
	_, ok := b.root.Get(k)
	return ok
}

// Build returns an immutable Map holding the builder's current root and
// permanently disables the builder: every method, including a repeated
// Build, panics with "immutable: builder used after Build" from this point
// on. Freeze invalidates the owner token in O(1); nodes may retain stale
// inactive stamps.
func (b *MapBuilder[K, V]) Build() Map[K, V] {
	freezeBuilderWindow(&b.win)
	root := b.root
	b.root = hamt.Root[K, V]{}
	return Map[K, V]{root: root}
}
