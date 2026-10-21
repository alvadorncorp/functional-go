package wbtree

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// Root is an immutable weight-balanced search tree ordered by its comparator.
//
// The zero value has no comparator: Len, Get, All and Delete treat it as
// empty, while Set and TransientSet panic. Use New for a writable empty root.
type Root[K, V any] struct {
	cmp  func(a, b K) int
	node *node[K, V]
}

// New returns an empty Root ordered by cmp. cmp must be a strict weak order
// returning a negative, zero, or positive result; New panics when it is nil.
func New[K, V any](cmp func(a, b K) int) Root[K, V] {
	if cmp == nil {
		panic("wbtree: nil comparator")
	}
	return Root[K, V]{cmp: cmp}
}

// HasComparator reports whether r carries a comparator (false only for the
// zero Root).
func (r Root[K, V]) HasComparator() bool { return r.cmp != nil }

// Len reports the number of entries in r.
func (r Root[K, V]) Len() int { return size(r.node) }

// Get returns the value stored for a key equivalent to key and true, or the
// zero V and false.
func (r Root[K, V]) Get(key K) (V, bool) {
	for n := r.node; n != nil; {
		switch c := r.cmp(key, n.key); {
		case c < 0:
			n = n.left
		case c > 0:
			n = n.right
		default:
			return n.val, true
		}
	}
	var zero V
	return zero, false
}

// Set returns a new Root with key mapped to val. r is unchanged. An
// equivalent stored key is replaced by key. New nodes are published with
// owner == nil (pure path-copy). Set panics on a zero Root.
func (r Root[K, V]) Set(key K, val V) Root[K, V] {
	return r.set(nil, key, val)
}

// Delete returns a new Root without key. r is unchanged. When key is absent,
// r itself is returned (same node pointer). An emptied Root keeps its
// comparator.
func (r Root[K, V]) Delete(key K) Root[K, V] {
	return r.delete(nil, key)
}

// TransientSet returns a Root with key mapped to val, editing under owner.
// Nodes exclusively owned by an active owner are mutated in place; others are
// shallow-cloned and stamped. An equivalent stored key is replaced by key.
// TransientSet panics on a zero Root.
func (r Root[K, V]) TransientSet(owner *ownership.Owner, key K, val V) Root[K, V] {
	return r.set(owner, key, val)
}

// TransientDelete returns a Root without key, editing under owner. When key
// is absent, r itself is returned (same node pointer).
func (r Root[K, V]) TransientDelete(owner *ownership.Owner, key K) Root[K, V] {
	return r.delete(owner, key)
}

// All yields every (key, value) pair in ascending comparator order.
func (r Root[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		r.node.forEach(yield)
	}
}

// set is the shared Set/TransientSet path; a nil owner path-copies.
func (r Root[K, V]) set(owner *ownership.Owner, key K, val V) Root[K, V] {
	if r.cmp == nil {
		panic("wbtree: no comparator")
	}
	return Root[K, V]{cmp: r.cmp, node: insert(owner, r.cmp, r.node, key, val)}
}

// delete is the shared Delete/TransientDelete path; a nil owner path-copies.
func (r Root[K, V]) delete(owner *ownership.Owner, key K) Root[K, V] {
	next, removed := remove(owner, r.cmp, r.node, key)
	if !removed {
		return r
	}
	return Root[K, V]{cmp: r.cmp, node: next}
}
