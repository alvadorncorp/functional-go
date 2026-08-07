package persistent

import (
	"iter"
	"maps"

	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// cloneMap returns a new, always non-nil map holding a copy of m, with
// spare capacity for extra additional entries. It is shared by Map and Set,
// whose write methods must never mutate the receiver's backing map nor hand
// out a nil map that would panic on the caller's first write into it.
func cloneMap[K comparable, V any](m map[K]V, extra int) map[K]V {
	next := make(map[K]V, len(m)+extra)
	for k, v := range m {
		next[k] = v
	}
	return next
}

// Map is an immutable key-value collection. Every write method returns a
// new Map; the receiver is never changed. See the package doc for the full
// contract, including why Map is not ==-comparable, why iteration order is
// unspecified, and why documented complexity is an upper bound rather than
// today's constant.
type Map[K comparable, V any] struct {
	items map[K]V
}

// NewMap returns a Map holding a copy of m. The caller's map is never
// retained: later mutation of m through the caller's own reference does not
// reach the returned Map.
func NewMap[K comparable, V any](m map[K]V) Map[K, V] {
	return Map[K, V]{items: cloneMap(m, 0)}
}

// CollectMap drains seq into a new Map. When seq yields the same key more
// than once, the last value for that key wins, matching maps.Collect.
func CollectMap[K comparable, V any](seq iter.Seq2[K, V]) Map[K, V] {
	return Map[K, V]{items: maps.Collect(seq)}
}

// Len returns the number of entries in m.
func (m Map[K, V]) Len() int { return len(m.items) }

// Get returns the value stored for k and true, or the zero value of V and
// false when k is absent.
func (m Map[K, V]) Get(k K) (V, bool) {
	v, ok := m.items[k]
	return v, ok
}

// Has reports whether k is present in m.
func (m Map[K, V]) Has(k K) bool {
	_, ok := m.items[k]
	return ok
}

// Set returns a new Map with k mapped to v, replacing any existing value for
// k. m is unchanged.
func (m Map[K, V]) Set(k K, v V) Map[K, V] {
	next := cloneMap(m.items, 1)
	next[k] = v
	return Map[K, V]{items: next}
}

// Delete returns a new Map without k. m is unchanged.
func (m Map[K, V]) Delete(k K) Map[K, V] {
	next := cloneMap(m.items, 0)
	delete(next, k)
	return Map[K, V]{items: next}
}

// All returns an iterator over m's (key, value) pairs. Iteration order is
// UNSPECIFIED and may change between releases.
func (m Map[K, V]) All() iter.Seq2[K, V] { return maps.All(m.items) }

// Keys returns an iterator over m's keys. Iteration order is UNSPECIFIED and
// may change between releases.
func (m Map[K, V]) Keys() iter.Seq[K] { return maps.Keys(m.items) }

// Values returns an iterator over m's values. Iteration order is
// UNSPECIFIED and may change between releases.
func (m Map[K, V]) Values() iter.Seq[V] { return maps.Values(m.items) }

// Items returns a new, non-nil map holding a copy of m's entries. The
// caller owns the result: mutating it never reaches m, and later writes to m
// never reach a map returned earlier.
func (m Map[K, V]) Items() map[K]V { return cloneMap(m.items, 0) }

// Equal reports whether m and other hold the same set of keys, each mapped
// to an equal value under functional/internal/eq.Equal.
func (m Map[K, V]) Equal(other Map[K, V]) bool {
	if len(m.items) != len(other.items) {
		return false
	}
	for k, v := range m.items {
		ov, ok := other.items[k]
		if !ok || !eq.Equal(v, ov) {
			return false
		}
	}
	return true
}

// Builder returns a MapBuilder seeded with a copy of m's entries. m is not
// affected by any later operation on the builder.
func (m Map[K, V]) Builder() *MapBuilder[K, V] {
	return &MapBuilder[K, V]{
		items: cloneMap(m.items, 0),
		win:   transient.Open(),
	}
}

// MapBuilder accumulates writes for Map in amortized linear time. It is
// MUTABLE and is NOT safe for concurrent use.
//
// Build transfers ownership of the builder's internal buffer to the Map it
// returns. After Build runs, every method on the builder — including Build
// itself — panics with "persistent: builder used after Build".
type MapBuilder[K comparable, V any] struct {
	items map[K]V
	win   *transient.Window
}

func (b *MapBuilder[K, V]) checkNotBuilt() {
	checkBuilderWindow(b.win)
}

// Set maps k to v in the builder's buffer, replacing any existing value for
// k, and returns b for chaining.
func (b *MapBuilder[K, V]) Set(k K, v V) *MapBuilder[K, V] {
	b.checkNotBuilt()
	b.items[k] = v
	return b
}

// Delete removes k from the builder's buffer and returns b for chaining.
func (b *MapBuilder[K, V]) Delete(k K) *MapBuilder[K, V] {
	b.checkNotBuilt()
	delete(b.items, k)
	return b
}

// Len returns the number of entries currently buffered.
func (b *MapBuilder[K, V]) Len() int {
	b.checkNotBuilt()
	return len(b.items)
}

// Build returns an immutable Map holding the builder's buffered entries and
// permanently disables the builder: every method, including a repeated
// Build, panics with "persistent: builder used after Build" from this point
// on.
func (b *MapBuilder[K, V]) Build() Map[K, V] {
	freezeBuilderWindow(&b.win)
	items := b.items
	b.items = nil
	return Map[K, V]{items: items}
}
