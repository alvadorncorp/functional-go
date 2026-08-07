package hamt

import (
	"hash/maphash"
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
	"github.com/alvadorncorp/gopackages/functional/internal/engine/keyhash"
)

// Root is an immutable hash-array-mapped trie of comparable keys.
// Each independent empty Root holds its own maphash.Seed; every derived
// version (Set/Delete) preserves that seed so structural sharing stays valid.
//
// Zero value is not a valid empty root (seed would be zero). Use Empty.
//
// TransientSet/TransientDelete edit under an active *ownership.Owner: owned
// nodes mutate in place; foreign or published nodes are shallow-cloned and
// stamped. After the owner is invalidated (Freeze), residual stamps are
// treated as published — no tree walk is required to publish.
type Root[K comparable, V any] struct {
	seed maphash.Seed
	node node[K, V]
	size int
}

// Empty returns an empty Root with a fresh maphash seed.
func Empty[K comparable, V any]() Root[K, V] {
	return Root[K, V]{seed: maphash.MakeSeed()}
}

// Len reports the number of entries in r.
func (r Root[K, V]) Len() int { return r.size }

// Get returns the value for key and true, or the zero V and false.
func (r Root[K, V]) Get(key K) (V, bool) {
	if r.node == nil {
		var zero V
		return zero, false
	}
	return r.node.get(keyhash.Hash(r.seed, key), 0, key)
}

// Set returns a new Root with key mapped to val. r is unchanged.
// The result preserves r's seed. Size grows only when key was absent.
// New nodes are published with owner == nil (pure path-copy).
func (r Root[K, V]) Set(key K, val V) Root[K, V] {
	h := keyhash.Hash(r.seed, key)
	if r.node == nil {
		return Root[K, V]{
			seed: r.seed,
			node: &leaf[K, V]{hash: h, key: key, val: val},
			size: 1,
		}
	}
	next, added := r.node.set(h, 0, key, val)
	size := r.size
	if added {
		size++
	}
	return Root[K, V]{seed: r.seed, node: next, size: size}
}

// Delete returns a new Root without key. r is unchanged.
// When key is absent, the same Root value is returned (same node pointer).
// Fully emptied roots still preserve the original seed.
// New nodes are published with owner == nil (pure path-copy).
func (r Root[K, V]) Delete(key K) Root[K, V] {
	if r.node == nil {
		return r
	}
	h := keyhash.Hash(r.seed, key)
	next, removed := r.node.del(h, 0, key)
	if !removed {
		return r
	}
	return Root[K, V]{seed: r.seed, node: next, size: r.size - 1}
}

// TransientSet returns a Root with key mapped to val, editing under owner.
// Nodes exclusively owned by an active owner are mutated in place; others are
// shallow-cloned and stamped. r's node graph is not mutated when r's nodes are
// published or owned by a different/inactive token.
func (r Root[K, V]) TransientSet(owner *ownership.Owner, key K, val V) Root[K, V] {
	h := keyhash.Hash(r.seed, key)
	if r.node == nil {
		return Root[K, V]{
			seed: r.seed,
			node: &leaf[K, V]{owner: owner, hash: h, key: key, val: val},
			size: 1,
		}
	}
	next, added := r.node.tset(owner, h, 0, key, val)
	size := r.size
	if added {
		size++
	}
	return Root[K, V]{seed: r.seed, node: next, size: size}
}

// TransientDelete returns a Root without key, editing under owner.
// When key is absent, the same Root value is returned (same node pointer).
func (r Root[K, V]) TransientDelete(owner *ownership.Owner, key K) Root[K, V] {
	if r.node == nil {
		return r
	}
	h := keyhash.Hash(r.seed, key)
	next, removed := r.node.tdel(owner, h, 0, key)
	if !removed {
		return r
	}
	return Root[K, V]{seed: r.seed, node: next, size: r.size - 1}
}

// All yields every (key, value) pair. Order is unspecified.
func (r Root[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if r.node != nil {
			r.node.forEach(yield)
		}
	}
}
