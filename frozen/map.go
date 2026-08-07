package frozen

import (
	"iter"
	"maps"

	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// Map is an immutable snapshot of a map[K]V, with the same copy-on-ingress
// and copy-on-egress contract as Slice. The zero Map is a valid, empty Map.
type Map[K comparable, V any] struct {
	items map[K]V
}

// NewMap copies m into a Map. Later writes to m cannot reach the Map.
func NewMap[K comparable, V any](m map[K]V) Map[K, V] {
	return Map[K, V]{items: maps.Clone(m)}
}

// CollectMap copies the key/value pairs produced by seq into a Map.
func CollectMap[K comparable, V any](seq iter.Seq2[K, V]) Map[K, V] {
	return Map[K, V]{items: maps.Collect(seq)}
}

// Get returns comma-ok, deliberately NOT functional/optional.Optional[V]: it
// is this choice that keeps frozen independent of optional (no public
// subpackage of functional imports another). Bridge at the call site with
// optional.FromOK(m.Get(k)).
func (m Map[K, V]) Get(k K) (V, bool) {
	v, ok := m.items[k]
	return v, ok
}

// Has reports whether k is present in m.
func (m Map[K, V]) Has(k K) bool {
	_, ok := m.items[k]
	return ok
}

// Len returns the number of key/value pairs in m.
func (m Map[K, V]) Len() int { return len(m.items) }

// All iterates key/value pairs in unspecified order, like a native Go map.
func (m Map[K, V]) All() iter.Seq2[K, V] {
	return maps.All(m.items)
}

// Keys iterates the keys in unspecified order.
func (m Map[K, V]) Keys() iter.Seq[K] {
	return maps.Keys(m.items)
}

// Values iterates the values in unspecified order.
func (m Map[K, V]) Values() iter.Seq[V] {
	return maps.Values(m.items)
}

// Items returns a fresh mutable copy of the map. Every call allocates.
func (m Map[K, V]) Items() map[K]V {
	return maps.Clone(m.items)
}

// Equal reports whether m and other hold the same key/value pairs, comparing
// values element-wise via functional/internal/eq.Equal. Order does not
// matter: m is a map.
func (m Map[K, V]) Equal(other Map[K, V]) bool {
	return MapEqualFunc(m, other, eq.Equal[V])
}
